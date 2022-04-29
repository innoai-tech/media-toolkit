package video

import (
	"context"
	"fmt"
	"github.com/containerd/containerd/content"
	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/mp4"
	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"io"
	"os"
	"strconv"
)

func newRecorder(ctx context.Context, s storage.Store, ref string, codecData []av.CodecData) (*recorder, error) {
	tmp, err := os.CreateTemp("", "video")
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, "create temp tmp failed")
		return nil, err
	}

	r := &recorder{
		s:     s,
		ref:   ref,
		tmp:   tmp,
		Muxer: mp4.NewMuxer(tmp),
	}

	if err := r.WriteHeader(codecData); err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, "muxer write header failed")
		return nil, err
	}
	return r, nil
}

type recorder struct {
	s   storage.Store
	ref string
	tmp *os.File
	*mp4.Muxer
}

func (r *recorder) Commit(ctx context.Context) error {
	defer func() {
		fmt.Println(r.tmp.Name())
		//_ = r.tmp.Truncate(0)
		_ = r.tmp.Close()
	}()

	if err := r.WriteTrailer(); err != nil {
		return err
	}

	cw, err := r.s.Writer(ctx, content.WithRef(r.ref))
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

	return cw.Commit(context.Background(), size, cw.Digest(), content.WithLabels(map[string]string{
		"$ref":      r.ref,
		"mediaType": storage.MediaTypeVideoMP4,
		"size":      strconv.Itoa(int(size)),
	}))
}
