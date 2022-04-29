package video

import (
	"context"
	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/mp4"
	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"github.com/innoai-tech/media-toolkit/pkg/storage/mime"
	"github.com/prometheus/common/model"
	"io"
	"os"
	"strconv"
	"time"
)

func newRecorder(ctx context.Context, s storage.Ingester, id string, from time.Time, codecData []av.CodecData) (*recorder, error) {
	tmp, err := os.CreateTemp("", "video")
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, "create temp tmp failed")
		return nil, err
	}

	r := &recorder{
		ingester: s,
		id:       id,
		from:     from,
		tmp:      tmp,
		Muxer:    mp4.NewMuxer(tmp),
	}

	if err := r.WriteHeader(codecData); err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, "muxer write header failed")
		return nil, err
	}
	return r, nil
}

type recorder struct {
	ingester storage.Ingester
	id       string
	from     time.Time
	tmp      *os.File
	*mp4.Muxer
}

func (r *recorder) Commit(ctx context.Context) error {
	defer func() {
		_ = r.tmp.Truncate(0)
		_ = r.tmp.Close()
	}()

	if err := r.WriteTrailer(); err != nil {
		return err
	}

	cw, err := r.ingester.Writer(ctx)
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, "create writer failed")
		return err
	}

	// mv to start for Read
	if _, err := r.tmp.Seek(0, 0); err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, "seek failed")
		return err
	}

	size, err := io.Copy(cw, r.tmp)
	if err != nil {
		return err
	}

	return cw.Commit(context.Background(), size, cw.Digest(),
		blob.WithFromThough(model.TimeFromUnixNano(r.from.UnixNano()), model.TimeFromUnixNano(time.Now().UnixNano())),
		blob.WithLabels(map[string][]string{
			"_media_type": {mime.MediaTypeVideoMP4},
			"_device_id":  {r.id},
			"_size":       {strconv.Itoa(int(size))},
		}))
}
