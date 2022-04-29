package video

import (
	"context"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"sync"
	"text/template"

	"github.com/deepch/vdk/av"
	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
)

type Options struct {
	Filename string
	Log      *logr.Logger
}

func New(ingester storage.Ingester, options Options) livestream.StreamObserver {
	t, err := template.New(options.Filename).Parse(options.Filename)
	if err != nil {
		panic(err)
	}

	return &videoObserver{
		ingester: ingester,
		options:  options,
		t:        t,
	}
}

type videoObserver struct {
	ingester  storage.Ingester
	t         *template.Template
	options   Options
	codecData []av.CodecData
	r         *recorder
	rw        sync.RWMutex
}

func (w *videoObserver) Log() logr.Logger {
	if w.options.Log == nil {
		return logr.Discard()
	}
	return *w.options.Log
}

func (w *videoObserver) SetCodecData(codecData []av.CodecData) {
	w.codecData = codecData
}

func (w *videoObserver) Close() error {
	if w.Enabled() {
		err := w.r.Commit(logr.NewContext(context.Background(), w.Log()))
		w.r = nil
		return err
	}
	return nil
}

func (w *videoObserver) Enabled() bool {
	w.rw.Lock()
	defer w.rw.Unlock()

	return w.r != nil
}

func (w *videoObserver) WritePacket(pkt livestream.Packet) {
	ctx := logr.NewContext(context.Background(), w.Log())

	if !w.Enabled() && pkt.IsKeyFrame {
		// start record should start with when first key frame
		r, err := newRecorder(ctx, w.ingester, pkt.ID, pkt.At, w.codecData)
		if err != nil {
			w.Log().Error(err, "start record failed")
			return
		}
		w.r = r
	}

	if err := w.r.WritePacket(pkt.Packet); err != nil {
		w.Log().Error(err, "write packet failed")
	}
}
