package image

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/format"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"github.com/innoai-tech/media-toolkit/pkg/storage/mime"
	"github.com/innoai-tech/media-toolkit/pkg/util/ioutil"
	"github.com/innoai-tech/media-toolkit/pkg/util/syncutil"
	"image/jpeg"
	"io"
	"time"
)

func New(s storage.Ingester) livestream.StreamObserver {
	return &imageWriter{
		s:             s,
		CloseNotifier: syncutil.NewCloseNotifier(),
	}
}

type imageWriter struct {
	syncutil.CloseNotifier
	s storage.Ingester
}

func (w *imageWriter) Name() string {
	return "Image"
}

func (w *imageWriter) OnVideoSource(ctx context.Context, videoSource livestream.VideoSource) {
	if err := w.takePic(videoSource); err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, "take pic failed")
	}
}

func (w *imageWriter) takePic(videoSource livestream.VideoSource) error {
	defer w.Shutdown(nil)

	r, err := videoSource.NewReader()
	if err != nil {
		return err
	}

	img, release, err := r.Read()
	if err != nil {
		return err
	}
	defer release()

	now := time.Now()

	return ioutil.Pipe(
		func(w io.Writer) error {
			return jpeg.Encode(w, img, nil)
		},
		func(r io.Reader) error {
			return format.CommitTo(context.Background(), r, w.s, format.Info{
				ID:        videoSource.ID(),
				MediaType: mime.MediaTypeImageJPEG,
				StartedAt: now,
				At:        now,
			})
		},
	)
}
