package image

import (
	"context"
	"github.com/innoai-tech/media-toolkit/pkg/util/ioutil"
	"image/jpeg"
	"io"
	"time"

	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/format"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"github.com/innoai-tech/media-toolkit/pkg/storage/mime"
	"github.com/innoai-tech/media-toolkit/pkg/util/syncutil"
	"github.com/pion/mediadevices"
)

func New(ctx context.Context, s storage.Ingester) livestream.StreamObserver {
	return &imageWriter{
		s:             s,
		l:             logr.FromContextOrDiscard(ctx),
		CloseNotifier: syncutil.NewCloseNotifier(),
	}
}

type imageWriter struct {
	syncutil.CloseNotifier

	l    logr.Logger
	s    storage.Ingester
	once syncutil.Once
}

func (w *imageWriter) Name() string {
	return "Image"
}

func (w *imageWriter) OnVideoSource(videoSource mediadevices.VideoSource) {
	err := w.once.Do(func() error {
		defer w.Shutdown(nil)

		img, release, err := videoSource.Read()
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
	})
	if err != nil {
		w.l.Error(err, "take pic failed")
	}
}
