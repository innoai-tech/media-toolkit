package video

import (
	"context"
	"github.com/innoai-tech/media-toolkit/pkg/format"
	"github.com/innoai-tech/media-toolkit/pkg/storage/mime"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"github.com/innoai-tech/media-toolkit/pkg/util/syncutil"
)

type Options struct {
	MaxDuration time.Duration
}

type OptFunc func(o *Options)

func (o *Options) Apply(opts ...OptFunc) {
	for i := range opts {
		opts[i](o)
	}
}

func New(ingester storage.Ingester, opts ...OptFunc) livestream.StreamObserver {
	options := &Options{
		MaxDuration: 60 * time.Second,
	}

	options.Apply(opts...)

	return &videoObserver{
		CloseNotifier: syncutil.NewCloseNotifier(),
		ingester:      ingester,
		options:       *options,
	}
}

type videoObserver struct {
	options  Options
	ingester storage.Ingester
	syncutil.CloseNotifier
}

func (o *videoObserver) Close() error {
	return o.CloseNotifier.Close()
}

func (o *videoObserver) Stop() error {
	o.CloseNotifier.Shutdown(nil)
	return nil
}

func (o *videoObserver) Name() string {
	return "Video"
}

func (o *videoObserver) OnVideoSource(ctx context.Context, videoSource livestream.VideoSource) {
	l := logr.FromContextOrDiscard(ctx)

	f, err := os.CreateTemp("", "video-")
	if err != nil {
		return
	}

	encodedReader, err := videoSource.NewEncodedReader(livestream.Preset1080P)
	if err != nil {
		return
	}

	r := format.NewRecorder(f, encodedReader, videoSource.ID())

	timer := time.NewTimer(o.options.MaxDuration)

	go func() {
		var info *format.Info

		// close temp file
		defer func() {
			_ = f.Truncate(0)
			_ = f.Close()
		}()
		// commit to ingres
		defer func() {
			if info == nil {
				return
			}

			err := format.CommitTo(logr.NewContext(context.Background(), l), f, o.ingester, format.Info{
				MediaType: mime.MediaTypeVideoMP4,
				ID:        info.ID,
				StartedAt: info.StartedAt,
				At:        info.At,
			})

			if err != nil {
				l.Error(err, "commit video failed")
			}
		}()
		// close recorder
		defer r.Close()

		for {
			select {
			case <-timer.C:
				return
			case <-o.Done():
				return
			default:
				i, err := r.Record()
				if err != nil {
					l.Error(err, "commit video failed")
					return
				}
				info = i
			}
		}
	}()
}
