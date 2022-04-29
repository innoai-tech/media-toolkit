package wsmp4f

import (
	"context"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/mp4f"
	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	"github.com/innoai-tech/media-toolkit/pkg/util/syncutil"
)

func New(ctx context.Context, ws *websocket.Conn) livestream.StreamObserver {
	return &wsmp4f{
		l:             logr.FromContextOrDiscard(ctx),
		ws:            ws,
		CloseNotifier: syncutil.NewCloseNotifier(),
	}
}

type wsmp4f struct {
	ws *websocket.Conn
	l  logr.Logger

	chPkt *syncutil.Chan[av.Packet]
	once  syncutil.Once
	syncutil.CloseNotifier
}

func (w *wsmp4f) Name() string {
	return "WsMP4f"
}

func (w *wsmp4f) Close() error {
	if w.once.Ready() {
		w.chPkt.Close()
		w.once.Reset()
	}
	_ = w.CloseNotifier.Close()
	return w.ws.Close()
}

func (w *wsmp4f) WritePacket(pkt *livestream.Packet) {
	if codecs := pkt.Codecs; len(codecs) > 0 {
		if pkt.IsKeyFrame {
			w.once.Do(func() {
				muxer := mp4f.NewMuxer(nil)
				_ = muxer.WriteHeader(codecs)
				_ = w.ws.SetWriteDeadline(time.Now().Add(5 * time.Second))

				meta, init := muxer.GetInit(codecs)
				_ = w.ws.WriteMessage(websocket.BinaryMessage, append([]byte{9}, meta...))
				_ = w.ws.WriteMessage(websocket.BinaryMessage, init)

				w.chPkt = syncutil.NewBufferChan[av.Packet](100)

				go func() {
					for p := range w.chPkt.Recv() {
						ready, buf, err := muxer.WritePacket(p, false)
						if err != nil {
							w.l.Error(err, "write packet failed")
							continue
						}
						if ready {
							if err := w.ws.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
								w.l.Error(err, "SetWriteDeadline")
								continue
							}
							if err := w.ws.WriteMessage(websocket.BinaryMessage, buf); err != nil {
								w.l.Error(err, "WriteMessage")
							}
						}
					}
				}()
			})
		}

		if w.once.Ready() {
			w.chPkt.Send(pkt.Packet)
		}
	}
}
