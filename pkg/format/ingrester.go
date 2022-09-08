package format

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"github.com/innoai-tech/media-toolkit/pkg/types"
	"io"
	"strconv"
)

func CommitTo(ctx context.Context, f io.Reader, store storage.Ingester, info Info) error {
	l := logr.FromContextOrDiscard(ctx)

	cw, err := store.Writer(ctx)
	if err != nil {
		l.Error(err, "create writer failed")
		return err
	}

	// mv to start for Read
	if s, ok := f.(io.Seeker); ok {
		if _, err := s.Seek(0, 0); err != nil {
			l.Error(err, "seek failed")
			return err
		}
	}

	size, err := io.Copy(cw, f)
	if err != nil {
		return err
	}

	return cw.Commit(ctx, size, cw.Info().Digest(),
		blob.WithFromThough(types.TimeFromUnixNano(info.StartedAt.UnixNano()), types.TimeFromUnixNano(info.At.UnixNano())),
		blob.WithLabels(map[string][]string{
			"_media_type": {info.MediaType},
			"_device_id":  {info.ID},
			"_size":       {strconv.Itoa(int(size))},
		}),
	)
}
