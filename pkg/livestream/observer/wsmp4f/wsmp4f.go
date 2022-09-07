package wsmp4f

import (
	"context"
	"encoding/json"
	"io"

	"github.com/deepch/vdk/format/mp4f"
	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	"github.com/innoai-tech/media-toolkit/pkg/util/syncutil"
)

func New(ctx context.Context, w io.Writer) livestream.StreamObserver {
	return &wsmp4f{
		l:             logr.FromContextOrDiscard(ctx),
		w:             w,
		CloseNotifier: syncutil.NewCloseNotifier(),
	}
}

type wsmp4f struct {
	w     io.Writer
	l     logr.Logger
	chPkt *syncutil.Chan[*livestream.Packet]
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
	return w.CloseNotifier.Close()
}

type BufType byte

const (
	BufTypeFrame    BufType = 0
	BufTypeMetadata BufType = 8
	BufTypeCodec    BufType = 9
)

func (m BufType) Build(b []byte) []byte {
	return append([]byte{byte(m)}, b...)
}

func (w *wsmp4f) WritePacket(pkt livestream.Packet) {
	if codecs := pkt.Codecs; len(codecs) > 0 {
		if pkt.IsKeyFrame {
			_ = w.once.Do(func() error {
				muxer := mp4f.NewMuxer(nil)

				if err := muxer.WriteHeader(codecs); err != nil {
					w.l.Error(err, "muxer.WriteHeader")
					return err
				}

				meta, init := muxer.GetInit(codecs)
				if _, err := w.w.Write(BufTypeCodec.Build([]byte(meta))); err != nil {
					w.l.Error(err, "write codec")
					return err
				}

				if _, err := w.w.Write(init); err != nil {
					w.l.Error(err, "write codec init")
					return err
				}

				w.chPkt = syncutil.NewChan[*livestream.Packet]()

				go func() {
					for p := range w.chPkt.Recv() {
						ready, buf, err := muxer.WritePacket(p.Packet, false)
						if err != nil {
							w.l.Error(err, "write packet failed")
						}
						if ready {
							metaBuf, err := json.Marshal(p.Metadata)
							if err != nil {
								w.l.Error(err, "Marshal metadata")
								return
							}
							if _, err := w.w.Write(BufTypeMetadata.Build(metaBuf)); err != nil {
								w.l.Error(err, "Write metadata")
								return
							}
							if _, err := w.w.Write(buf); err != nil {
								w.l.Error(err, "Write frame")
								return
							}
						}
					}
				}()

				return nil
			})
		}

		if w.once.Ready() {
			w.chPkt.Send(&pkt)
		}
	}
}
