package local

import (
	"context"
	"fmt"
	"github.com/containerd/containerd/errdefs"
	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/storage/content"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		buffer := make([]byte, 1<<20)
		return &buffer
	},
}

func NewStore(root string) (content.Store, error) {
	if err := os.MkdirAll(filepath.Join(root, "ingest"), 0777); err != nil {
		return nil, err
	}
	return &contentStore{root: root}, nil
}

type contentStore struct {
	root string
}

func (s *contentStore) Delete(ctx context.Context, ref blob.Ref) error {
	bp := ref.BlobPath(s.root)

	if err := os.RemoveAll(bp); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("content: %w", ErrNotFound)
	}
	return nil
}

func (s *contentStore) Writer(ctx context.Context, opts ...blob.Opt) (content.Writer, error) {
	info := &blob.Info{}
	for i := range opts {
		opts[i](info)
	}

	if int64(info.From) == 0 {
		info.From = model.Now()
	}

	w, err := s.writer(ctx, info)
	if err != nil {
		return nil, err
	}
	return w, nil // lock is now held by w.
}

func (s *contentStore) ingestRoot(info *blob.Info) string {
	return filepath.Join(s.root, "ingest", digest.FromString(info.ExternalKey("")).Hex())
}

func (s *contentStore) writer(ctx context.Context, info *blob.Info) (content.Writer, error) {
	if info.Hex != "" {
		p := info.BlobPath(s.root)
		if _, err := os.Stat(p); err == nil {
			return nil, fmt.Errorf("content %v: %w", info.Hex, ErrAlreadyExists)
		}
	}

	root := s.ingestRoot(info)
	dataFile := filepath.Join(root, "data")

	var (
		digester  = digest.Canonical.Digester()
		offset    int64
		startedAt time.Time
		updatedAt time.Time
	)

	if err := os.Mkdir(root, 0755); err != nil {
		if !os.IsExist(err) {
			return nil, err
		}
	}

	fp, err := os.OpenFile(dataFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open data file: %w", err)
	}

	return &writer{
		s:         s,
		file:      fp,
		info:      info,
		path:      root,
		offset:    offset,
		digester:  digester,
		startedAt: startedAt,
		updatedAt: updatedAt,
	}, nil
}

type writer struct {
	s         *contentStore
	info      *blob.Info
	file      *os.File
	path      string // path to writer dir
	offset    int64
	total     int64
	digester  digest.Digester
	startedAt time.Time
	updatedAt time.Time
}

func (w *writer) Status() (content.Status, error) {
	return content.Status{
		Offset:    w.offset,
		Total:     w.total,
		StartedAt: w.startedAt,
		UpdatedAt: w.updatedAt,
	}, nil
}

func (w *writer) Info() blob.Info {
	return *w.info
}

// Write p to the transaction.
//
// Note that writes are unbuffered to the backing file. When writing, it is
// recommended to wrap in a bufio.Writer or, preferably, use io.CopyBuffer.
func (w *writer) Write(p []byte) (n int, err error) {
	n, err = w.file.Write(p)
	w.digester.Hash().Write(p[:n])
	w.offset += int64(len(p))
	w.updatedAt = time.Now()
	return n, err
}

func (w *writer) Commit(ctx context.Context, size int64, expected digest.Digest, opts ...blob.Opt) error {
	for _, opt := range opts {
		opt(w.info)
	}
	if w.info.UserID == "" {
		w.info.UserID = blob.DefaultUser
	}
	if w.info.Through == 0 {
		w.info.Through = w.info.From
	}

	file := w.file
	w.file = nil

	if file == nil {
		return fmt.Errorf("cannot commit on closed writer: %w", errdefs.ErrFailedPrecondition)
	}

	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync failed: %w", err)
	}

	fi, err := file.Stat()
	closeErr := file.Close()
	if err != nil {
		return fmt.Errorf("stat on ingest file failed: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("failed to close ingest file: %w", closeErr)
	}

	if size > 0 && size != fi.Size() {
		return fmt.Errorf("unexpected commit size %d, expected %d: %w", fi.Size(), size, errdefs.ErrFailedPrecondition)
	}

	dgst := w.digester.Digest()
	if expected != "" && expected != dgst {
		return fmt.Errorf("unexpected commit digest %s, expected %s: %w", dgst, expected, errdefs.ErrFailedPrecondition)
	}

	w.info.Alg = string(dgst.Algorithm())
	w.info.Hex = dgst.Hex()

	var (
		ingest = filepath.Join(w.path, "data")
		target = w.info.BlobPath(w.s.root)
	)

	// make sure parent directories of blob exist
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	if err := os.Rename(ingest, target); err != nil {
		return err
	}

	if err := os.RemoveAll(w.path); err != nil {
		logr.FromContextOrDiscard(ctx).WithValues("path", w.path).Error(errors.New("failed to remove ingest directory"), "")
	}

	return nil
}

func (w *writer) Close() (err error) {
	if w.file != nil {
		_ = w.file.Sync()
		err = w.file.Close()
		w.file = nil
		return
	}

	return nil
}

func (w *writer) Truncate(size int64) error {
	if size != 0 {
		return errors.New("Truncate: unsupported size")
	}
	w.offset = 0
	w.digester.Hash().Reset()
	if _, err := w.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return w.file.Truncate(0)
}
