package local

import (
	"context"
	"github.com/go-logr/logr"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/pkg/errors"

	"github.com/innoai-tech/media-toolkit/pkg/storage/label/index"
	"github.com/innoai-tech/media-toolkit/pkg/storage/label/util"
)

var (
	ErrUnexistentDB = errors.New("db file does not exist")
)

const (
	dbReloadPeriod = 10 * time.Minute
)

const (
	DBOperationRead = iota
	DBOperationWrite
)

type DB = pebble.DB

// DBConfig for a BoltDB index client.
type DBConfig struct {
	Directory string `yaml:"directory"`
}

// NewIndexClient creates a new indexClient that used BoltDB.
func NewIndexClient(cfg DBConfig) (index.Client, error) {
	root := filepath.Join(cfg.Directory, "labels")

	if err := util.EnsureDirectory(root); err != nil {
		return nil, err
	}

	ic := &indexClient{
		root: root,
		dbs:  map[string]*pebble.DB{},
		done: make(chan struct{}),
	}

	ic.wait.Add(1)

	go ic.loop()

	return ic, nil
}

type indexClient struct {
	root   string
	dbsMtx sync.RWMutex
	dbs    map[string]*pebble.DB
	wait   sync.WaitGroup
	done   chan struct{}
}

func (b *indexClient) loop() {
	defer b.wait.Done()

	ticker := time.NewTicker(dbReloadPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.reload()
		case <-b.done:
			return
		}
	}
}

func (b *indexClient) reload() {
	b.dbsMtx.RLock()

	var removedDBs []string
	for name := range b.dbs {
		if _, err := os.Stat(path.Join(b.root, name)); err != nil && os.IsNotExist(err) {
			removedDBs = append(removedDBs, name)
			continue
		}
	}

	b.dbsMtx.RUnlock()

	if len(removedDBs) != 0 {
		b.dbsMtx.Lock()
		defer b.dbsMtx.Unlock()

		for _, name := range removedDBs {
			if err := b.dbs[name].Close(); err != nil {
				continue
			}
			delete(b.dbs, name)
		}
	}
}

func (b *indexClient) Shutdown(ctx context.Context) error {
	close(b.done)

	b.dbsMtx.Lock()
	defer b.dbsMtx.Unlock()

	wg := sync.WaitGroup{}

	for _, db := range b.dbs {
		wg.Add(1)

		go func(db *pebble.DB) {
			defer wg.Done()

			defer func() {
				if err := db.Close(); err != nil {
					logr.FromContextOrDiscard(ctx).Error(err, "Close")
				}
			}()

			f, err := db.AsyncFlush()
			if err != nil {
				logr.FromContextOrDiscard(ctx).Error(err, "AsyncFlush")
				return
			}

			for {
				select {
				case <-ctx.Done():
					return
				case <-f:
					logr.FromContextOrDiscard(ctx).Info("Flushed")
					return
				}
			}
		}(db)
	}

	wg.Wait()
	return nil
}

func (b *indexClient) NewWriteBatch() index.WriteBatch {
	return &WriteBatch{
		Writes: map[string]*TableWrites{},
	}
}

// GetDB should always return a db for write operation unless an error occurs while doing so.
// While for read operation it should throw ErrUnexistentBoltDB error if file does not exist for reading
func (b *indexClient) GetDB(name string, operation int) (*pebble.DB, error) {
	b.dbsMtx.RLock()
	db, ok := b.dbs[name]
	b.dbsMtx.RUnlock()
	if ok {
		return db, nil
	}

	// we do not want to create a new db for reading if it does not exist
	if operation == DBOperationRead {
		if _, err := os.Stat(path.Join(b.root, name)); err != nil {
			if os.IsNotExist(err) {
				return nil, ErrUnexistentDB
			}
			return nil, err
		}
	}

	b.dbsMtx.Lock()
	defer b.dbsMtx.Unlock()
	db, ok = b.dbs[name]
	if ok {
		return db, nil
	}

	// Open the database.
	// Set Timeout to avoid obtaining file lock wait indefinitely.
	db, err := pebble.Open(path.Join(b.root, name), &pebble.Options{
		// https://github.com/cockroachdb/pebble/issues/1068#issuecomment-784208214
		L0CompactionThreshold: 2,
		L0StopWritesThreshold: 1000,
		LBaseMaxBytes:         64 << 20, // 64 MB
		MaxConcurrentCompactions: func() int {
			return 3
		},
		MemTableSize:                64 << 20, // 64 MB
		MemTableStopWritesThreshold: 4,
	})
	if err != nil {
		return nil, err
	}

	b.dbs[name] = db
	return db, nil
}

func (b *indexClient) WriteToDB(ctx context.Context, db *pebble.DB, writes *TableWrites) (e error) {
	batch := db.NewBatch()

	for i := range writes.puts {
		e := writes.puts[i]
		if err := batch.Set(e.Key(), e.Value, pebble.NoSync); err != nil {
			return err
		}
	}

	for i := range writes.deletes {
		e := writes.deletes[i]
		if err := batch.Delete(e.Key(), pebble.NoSync); err != nil {
			return err
		}
	}

	return batch.Commit(pebble.NoSync)
}

func (b *indexClient) BatchWrite(ctx context.Context, batch index.WriteBatch) error {
	writes := batch.(*WriteBatch).Writes
	for tableName := range writes {
		db, err := b.GetDB(tableName, DBOperationWrite)
		if err != nil {
			return err
		}
		if err := b.WriteToDB(ctx, db, writes[tableName]); err != nil {
			return err
		}
	}
	return nil
}

func (b *indexClient) QueryPages(ctx context.Context, queries []index.Query, callback index.QueryPagesCallback) error {
	return util.DoParallelQueries(ctx, b.query, queries, callback)
}

func (b *indexClient) query(ctx context.Context, query index.Query, callback index.QueryPagesCallback) error {
	db, err := b.GetDB(query.TableName, DBOperationRead)
	if err != nil {
		if err == ErrUnexistentDB {
			return nil
		}
		return err
	}
	return b.QueryDB(ctx, db, query, callback)
}

func (b *indexClient) QueryDB(ctx context.Context, db *pebble.DB, query index.Query, action index.QueryPagesCallback) error {
	return b.iteratorDB(ctx, db.NewIter(&pebble.IterOptions{}), query, action)
}

func (b *indexClient) iteratorDB(ctx context.Context, iter *pebble.Iterator, query index.Query, action index.QueryPagesCallback) error {
	batch := batchPool.Get().(*cursorBatch)
	defer batchPool.Put(batch)
	batch.reset(iter, &query)
	return action(batch, query)
}
