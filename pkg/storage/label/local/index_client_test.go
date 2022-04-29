package local

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/innoai-tech/media-toolkit/pkg/storage/label/index"

	. "github.com/octohelm/x/testing"
)

var (
	testKey   = []byte("test-key")
	testValue = []byte("test-value")
)

func TestDBWR(t *testing.T) {
	workingDir := t.TempDir()

	ic, err := NewIndexClient(DBConfig{
		Directory: workingDir,
	})

	Expect(t, err, Be[error](nil))
	defer ic.Stop()

	entry := index.Entry{
		TableName:  "test",
		HashValue:  "test",
		RangeValue: []byte(strconv.Itoa(2)),
	}

	t.Run("Write", func(t *testing.T) {
		w := ic.NewWriteBatch()

		for i := range make([]byte, 10) {
			w.Add(index.Entry{
				TableName:  "test",
				HashValue:  "test",
				RangeValue: []byte(strconv.Itoa(i)),
				Value:      []byte(strconv.Itoa(i)),
			})
		}

		err := ic.BatchWrite(context.Background(), w)
		Expect(t, err, Be[error](nil))

		db, err := ic.(*indexClient).GetDB("test", DBOperationRead)
		Expect(t, err, Be[error](nil))
		_, _, err = db.Get(entry.Key())
		Expect(t, err, Be[error](nil))
	})

	t.Run("Delete", func(t *testing.T) {
		w := ic.NewWriteBatch()
		w.Delete(index.Entry{
			TableName:  "test",
			HashValue:  "test",
			RangeValue: []byte(strconv.Itoa(2)),
		})
		err := ic.BatchWrite(context.Background(), w)
		Expect(t, err, Be[error](nil))

		db, err := ic.(*indexClient).GetDB("test", DBOperationRead)
		Expect(t, err, Be[error](nil))
		_, _, err = db.Get(entry.Key())
		Expect(t, err, Not(Be[error](nil)))
	})
}

func TestDBReload(t *testing.T) {
	workingDir := t.TempDir()

	ic, err := NewIndexClient(DBConfig{
		Directory: workingDir,
	})
	Expect(t, err, Be[error](nil))
	defer ic.Stop()

	t.Run("Given dbs", func(t *testing.T) {
		testDb1 := "test1"
		testDb2 := "test2"

		setupDB(t, ic.(*indexClient), testDb1)
		setupDB(t, ic.(*indexClient), testDb2)

		t.Run("When reload, will got two dbs", func(t *testing.T) {
			ic.(*indexClient).reload()
			Expect(t, len(ic.(*indexClient).dbs), Be[int](2))
		})

		t.Run("When remove one", func(t *testing.T) {
			err := os.RemoveAll(filepath.Join(workingDir, "labels", testDb1))
			Expect(t, err, Be[error](nil))

			droppedDb, err := ic.(*indexClient).GetDB(testDb1, DBOperationRead)
			Expect(t, err, Be[error](nil))

			valueFromDb, c, err := droppedDb.Get(testKey)
			Expect(t, err, Be[error](nil))
			_ = c.Close()

			Expect(t, valueFromDb, Equal(testValue))

			ic.(*indexClient).reload()
			Expect(t, len(ic.(*indexClient).dbs), Be(1))
			_, err = ic.(*indexClient).GetDB(testDb1, DBOperationRead)
			Expect(t, err, Be(ErrUnexistentDB))
		})
	})
}

func TestBoltDB_GetDB(t *testing.T) {
	workingDir := t.TempDir()
	ic, err := NewIndexClient(DBConfig{
		Directory: workingDir,
	})
	Expect(t, err, Be[error](nil))

	// setup a db to already exist
	testDb1 := "test1"
	setupDB(t, ic.(*indexClient), testDb1)

	// check whether an existing db can be fetched for reading
	_, err = ic.(*indexClient).GetDB(testDb1, DBOperationRead)
	Expect(t, err, Be[error](nil))

	// check whether read operation throws ErrUnexistentBoltDB error for db which does not exists
	unexistentDb := "unexistent-db"

	_, err = ic.(*indexClient).GetDB(unexistentDb, DBOperationRead)
	Expect(t, err, Be(ErrUnexistentDB))

	// check whether write operation sets up a new db for writing
	db, err := ic.(*indexClient).GetDB(unexistentDb, DBOperationWrite)
	Expect(t, err, Be[error](nil))
	Expect(t, db, Not(Be[*DB](nil)))

	// recreate index client to check whether we can read already created test1 db without writing first
	ic.Stop()
	ic, err = NewIndexClient(DBConfig{
		Directory: workingDir,
	})
	Expect(t, err, Be[error](nil))
	defer ic.Stop()

	_, err = ic.(*indexClient).GetDB(testDb1, DBOperationRead)
	Expect(t, err, Be[error](nil))
}

func setupDB(t *testing.T, boltdbIndexClient *indexClient, dbname string) {
	db, err := boltdbIndexClient.GetDB(dbname, DBOperationWrite)
	Expect(t, err, Be[error](nil))
	err = db.Set(testKey, testValue, pebble.Sync)
	Expect(t, err, Be[error](nil))
}
