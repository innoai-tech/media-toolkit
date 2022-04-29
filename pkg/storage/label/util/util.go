package util

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/innoai-tech/media-toolkit/pkg/storage/label/index"
)

type DoSingleQuery func(context.Context, index.Query, index.QueryPagesCallback) error

var QueryParallelism = 100

func DoParallelQueries(ctx context.Context, doSingleQuery DoSingleQuery, queries []index.Query, callback index.QueryPagesCallback) error {
	if len(queries) == 1 {
		return doSingleQuery(ctx, queries[0], callback)
	}

	queue := make(chan index.Query)
	incomingErrors := make(chan error)
	n := min(len(queries), QueryParallelism)

	// Run n parallel goroutines fetching queries from the queue
	for i := 0; i < n; i++ {
		go func() {
			// TODO add otel
			for {
				query, ok := <-queue
				if !ok {
					return
				}
				incomingErrors <- doSingleQuery(ctx, query, callback)
			}
		}()
	}

	// Send all the queries into the queue
	go func() {
		for _, query := range queries {
			queue <- query
		}
		close(queue)
	}()

	// Now receive all the results.
	var lastErr error
	for i := 0; i < len(queries); i++ {
		err := <-incomingErrors
		if err != nil {
			lastErr = err
		}
	}

	return lastErr
}

func EnsureDirectory(dir string) error {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0o777)
	} else if err == nil && !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}
	return err
}

type ReadCloserWithContextCancelFunc struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func NewReadCloserWithContextCancelFunc(readCloser io.ReadCloser, cancel context.CancelFunc) io.ReadCloser {
	return ReadCloserWithContextCancelFunc{
		ReadCloser: readCloser,
		cancel:     cancel,
	}
}

func (r ReadCloserWithContextCancelFunc) Close() error {
	defer r.cancel()
	return r.ReadCloser.Close()
}

func min[T int | float64](a, b T) T {
	if a < b {
		return a
	}
	return b
}
