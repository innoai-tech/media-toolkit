package local

import (
	"context"
	"fmt"
	"os"

	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/storage/content"
	"github.com/pkg/errors"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("not found")
)

func (s *contentStore) ReaderAt(ctx context.Context, ref blob.Ref) (content.ReaderAt, error) {
	blobPath := ref.BlobPath(s.root)
	reader, err := OpenReader(blobPath)
	if err != nil {
		return nil, fmt.Errorf("blob expected at %s: %w", blobPath, err)
	}
	return reader, nil
}

// OpenReader creates ReaderAt from a file
func OpenReader(p string) (content.ReaderAt, error) {
	fi, err := os.Stat(p)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		return nil, fmt.Errorf("blob not found: %w", ErrNotFound)
	}

	fp, err := os.Open(p)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		return nil, fmt.Errorf("blob not found: %w", ErrNotFound)
	}

	return sizeReaderAt{size: fi.Size(), fp: fp}, nil
}

type sizeReaderAt struct {
	size int64
	fp   *os.File
}

func (ra sizeReaderAt) ReadAt(p []byte, offset int64) (int, error) {
	return ra.fp.ReadAt(p, offset)
}

func (ra sizeReaderAt) Size() int64 {
	return ra.size
}

func (ra sizeReaderAt) Close() error {
	return ra.fp.Close()
}
