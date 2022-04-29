package content

import (
	"context"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/opencontainers/go-digest"
	"io"
	"time"
)

type Store interface {
	Remover
	Provider
	Ingester
}

type Remover interface {
	Delete(ctx context.Context, ref blob.Ref) error
}

type Provider interface {
	ReaderAt(ctx context.Context, ref blob.Ref) (ReaderAt, error)
}

type Ingester interface {
	Writer(ctx context.Context, opts ...blob.Opt) (Writer, error)
}

type ReaderAt interface {
	io.ReaderAt
	io.Closer
	Size() int64
}

type Writer interface {
	io.WriteCloser
	Commit(ctx context.Context, size int64, expected digest.Digest, opts ...blob.Opt) error
	Info() blob.Info
	Status() (Status, error)
	Truncate(size int64) error
}

type Status struct {
	Offset    int64
	Total     int64
	Expected  digest.Digest
	StartedAt time.Time
	UpdatedAt time.Time
}
