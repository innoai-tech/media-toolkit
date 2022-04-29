package video

import (
	"context"
	"github.com/deepch/vdk/av"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"github.com/innoai-tech/media-toolkit/pkg/util/syncutil"
	"time"

	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
)

type Options struct {
	MaxDuration time.Duration
}

type OptFunc func(o *Options)

func New(ctx context.Context, ingester storage.Ingester, opts ...OptFunc) livestream.StreamObserver {
	options := &Options{
		MaxDuration: 60 * time.Second,
	}

	for i := range opts {
		opts[i](options)
	}

	return &videoObserver{
		l:             logr.FromContextOrDiscard(ctx),
		ingester:      ingester,
		options:       *options,
		CloseNotifier: syncutil.NewCloseNotifier(),
	}
}

type videoObserver struct {
	options  Options
	l        logr.Logger
	ingester storage.Ingester
	recorder syncutil.ValueMutex[*recorder]
	syncutil.CloseNotifier
	chPkt *syncutil.Chan[av.Packet]
}

func (o *videoObserver) Name() string {
	return "Video"
}

func (o *videoObserver) Close() error {
	_ = o.CloseNotifier.Close()
	return o.Stop()
}

func (o *videoObserver) Stop() error {
	if r := o.recorder.Get(); r != nil {
		o.CloseNotifier.SendDone(nil)
		o.chPkt.Close()
		err := r.Commit(logr.NewContext(context.Background(), o.l))
		o.recorder.Set(nil)
		return err
	}
	return nil
}

func (o *videoObserver) WritePacket(pkt *livestream.Packet) {
	ctx := logr.NewContext(context.Background(), o.l)

	if r := o.recorder.Get(); r == nil && pkt.IsKeyFrame {
		// start record should start with when first key frame
		r, err := newRecorder(ctx, o.ingester, pkt.ID, pkt.At, pkt.Codecs)
		if err != nil {
			o.l.Error(err, "start record failed")
			return
		}

		o.recorder.Set(r)

		o.chPkt = syncutil.NewChan[av.Packet]()

		go func() {
			for p := range o.chPkt.Recv() {
				if err := r.WritePacket(p); err != nil {
					o.l.Error(err, "write packet failed")
				}
			}
		}()

		timer := time.NewTimer(o.options.MaxDuration)

		go func() {
			<-timer.C
			o.l.Info("auto stop...")
			if err := o.Stop(); err != nil {
				o.l.Error(err, "stop")
			}
		}()
	}

	if r := o.recorder.Get(); r != nil {
		o.chPkt.Send(pkt.Packet)
	}
}
