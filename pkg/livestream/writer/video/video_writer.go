package video

import (
	"bytes"
	"context"
	"fmt"
	"github.com/deepch/vdk/av"
	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"text/template"
)

type Options struct {
	Filename string
}

func NewStreamWriter(s storage.Store, options Options) livestream.StreamWriter {
	t, err := template.New(options.Filename).Parse(options.Filename)
	if err != nil {
		panic(err)
	}

	return &videoWriter{
		s:       s,
		options: options,
		t:       t,
	}
}

type videoWriter struct {
	s         storage.Store
	t         *template.Template
	options   Options
	codecData []av.CodecData
	r         *recorder
	lastPkt   *livestream.Packet
}

func (w *videoWriter) SetCodecData(codecData []av.CodecData) {
	w.codecData = codecData
}

func (w *videoWriter) Close() error {
	if w.Enabled() {
		w.Disable(context.Background())
	}
	return nil
}

func (w *videoWriter) Enable(ctx context.Context) {
	if w.r == nil && w.lastPkt != nil {
		ref := w.FrameFilename(w.lastPkt)
		r, err := newRecorder(ctx, w.s, ref, w.codecData)
		if err != nil {
			logr.FromContextOrDiscard(ctx).Error(err, "write packet failed")
			return
		}
		w.r = r
		if err := w.r.WritePacket(w.lastPkt.Packet); err != nil {
			logr.FromContextOrDiscard(ctx).Error(err, "write packet failed")
		}
	}
}

func (w *videoWriter) Enabled() bool {
	return w.r != nil
}

func (w *videoWriter) Disable(ctx context.Context) {
	if w.r != nil {
		_ = w.r.Commit(ctx)
		w.r = nil
	}
}

func (w *videoWriter) WritePacket(pkt *livestream.Packet) {
	ctx := pkt.Context()

	if w.r != nil {
		if err := w.r.WritePacket(pkt.Packet); err != nil {
			logr.FromContextOrDiscard(ctx).Error(err, "write packet failed")
		}
	}

	w.lastPkt = pkt
	w.Enable(ctx)
}

func (w *videoWriter) FrameFilename(pkt *livestream.Packet) string {
	buf := bytes.NewBuffer(nil)
	_ = w.t.Execute(buf, map[string]any{
		"Timestamp": pkt.At.Unix(),
	})
	return fmt.Sprintf("%s/%s", pkt.ID, buf.String())
}
