package wsmp4f

import (
	"context"
	"encoding/json"
	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/deepch/vdk/format/mp4f"
	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/format"
	"io"

	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	"github.com/innoai-tech/media-toolkit/pkg/util/syncutil"
)

func New(w io.Writer) livestream.StreamObserver {
	return &wsmp4f{
		w:             w,
		CloseNotifier: syncutil.NewCloseNotifier(),
	}
}

type wsmp4f struct {
	w io.Writer
	syncutil.CloseNotifier
	once syncutil.Once
}

func (w *wsmp4f) Name() string {
	return "WsMP4f"
}

func (w *wsmp4f) Close() error {
	if w.once.Ready() {
		w.once.Reset()
	}
	return w.CloseNotifier.Close()
}

func (w *wsmp4f) OnVideoSource(ctx context.Context, videoSource livestream.VideoSource) {
	_ = w.once.Do(func() error {
		encodedReader, err := videoSource.NewEncodedReader(livestream.Preset1080P)
		if err != nil {
			return err
		}

		go func() {
			defer encodedReader.Close()

			p := format.Packetizer{}
			muxer := mp4f.NewMuxer(nil)

			l := logr.FromContextOrDiscard(ctx)

			init := false

			for {
				select {
				case <-w.Done():
					return
				default:
					buf, _, err := encodedReader.Read()
					if err != nil {
						return
					}

					pkt := p.Packetize(buf.Data, buf.Samples)

					if !init {
						if !pkt.IsKeyFrame {
							continue
						}

						init = true

						codecData, err := h264parser.NewCodecDataFromSPSAndPPS(pkt.SPS, pkt.PPS)
						if err != nil {
							return
						}

						if err := muxer.WriteHeader([]av.CodecData{codecData}); err != nil {
							l.Error(err, "muxer.WriteHeader")
							return
						}

						meta, init := muxer.GetInit([]av.CodecData{codecData})
						if _, err := w.w.Write(BufTypeCodec.Build([]byte(meta))); err != nil {
							l.Error(err, "write codec")
							return
						}

						if _, err := w.w.Write(init); err != nil {
							l.Error(err, "write codec init")
							return
						}
					}

					if init {
						ready, b, err := muxer.WritePacket(av.Packet{
							Idx:        int8(pkt.Idx),
							IsKeyFrame: pkt.IsKeyFrame,
							Data:       pkt.Data,
							Time:       pkt.Time,
						}, false)
						if err != nil {
							l.Error(err, "write packet failed")
						}
						if ready {
							metaBuf, err := json.Marshal(livestream.Metadata{
								ID:        videoSource.ID(),
								At:        pkt.At,
								Observers: videoSource.Status().Observers,
							})
							if err != nil {
								l.Error(err, "Marshal metadata")
								return
							}
							if _, err := w.w.Write(BufTypeMetadata.Build(metaBuf)); err != nil {
								l.Error(err, "Write metadata")
								return
							}
							if _, err := w.w.Write(b); err != nil {
								l.Error(err, "Write frame")
								return
							}
						}
					}
				}
			}
		}()
		return nil
	})
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
