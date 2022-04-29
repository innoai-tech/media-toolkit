package wsmp4f

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/mp4f"
	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
)

var (
	writeWait = 5 * time.Second
)

func New(ctx context.Context, ws *websocket.Conn) livestream.StreamObserver {
	return &wsmp4f{
		l:     logr.FromContextOrDiscard(ctx),
		ws:    ws,
		chPkt: make(chan av.Packet, 100),
	}
}

type wsmp4f struct {
	l       logr.Logger
	ws      *websocket.Conn
	muxer   *mp4f.Muxer
	chPkt   chan av.Packet
	started int64
}

func (w *wsmp4f) Close() error {
	close(w.chPkt)
	return w.ws.Close()
}

func (w *wsmp4f) SetCodecData(codecs []av.CodecData) {
	go func() {
		timeLine := make(map[int8]time.Duration)

		for pkt := range w.chPkt {
			timeLine[pkt.Idx] += pkt.Duration
			pkt.Time = timeLine[pkt.Idx]

			ready, buf, err := w.muxer.WritePacket(pkt, false)
			if err != nil {
				w.l.Error(err, "write packet failed")
				continue
			}
			if ready {
				_ = w.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
				_ = w.ws.WriteMessage(websocket.BinaryMessage, buf)
			}
		}
	}()

	muxer := mp4f.NewMuxer(nil)
	_ = muxer.WriteHeader(codecs)
	w.muxer = muxer

	_ = w.ws.SetWriteDeadline(time.Now().Add(5 * time.Second))

	meta, init := muxer.GetInit(codecs)
	_ = w.ws.WriteMessage(websocket.BinaryMessage, append([]byte{9}, meta...))
	_ = w.ws.WriteMessage(websocket.BinaryMessage, init)
}

func (w *wsmp4f) WritePacket(pkt livestream.Packet) {
	defer func() {
		if e := recover(); e != nil {
		}
	}()

	if w.muxer != nil {
		if pkt.IsKeyFrame {
			atomic.AddInt64(&w.started, 1)
		}
		if atomic.LoadInt64(&w.started) > 0 {
			w.chPkt <- pkt.Packet
		}
	}
}
