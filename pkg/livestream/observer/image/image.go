package image

import (
	"bytes"
	"context"
	"fmt"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"github.com/innoai-tech/media-toolkit/pkg/storage/mime"
	"github.com/prometheus/common/model"
	"image/jpeg"
	"io"
	"strconv"
	"text/template"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/codec"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
)

type Options struct {
	Filename string
}

func New(s storage.Ingester, options Options) livestream.StreamObserver {
	t, err := template.New(options.Filename).Parse(options.Filename)
	if err != nil {
		panic(err)
	}

	return &imageWriter{
		options: options,
		s:       s,
		t:       t,
	}
}

type imageWriter struct {
	s       storage.Ingester
	t       *template.Template
	options Options
	decoder *codec.H264Decoder
}

func (w *imageWriter) SetCodecData(codecData []av.CodecData) {
	for _, c := range codecData {
		if cd, ok := c.(h264parser.CodecData); ok {
			w.decoder, _ = codec.NewH264Decoder(cd.Record)
		}
	}
}

func (w *imageWriter) Close() error {
	return w.decoder.Close()
}

func (w *imageWriter) WritePacket(pkt livestream.Packet) {
	if w.decoder != nil && pkt.IsKeyFrame {
		img, err := w.decoder.Decode(pkt.Data)
		if err == nil && img != nil {
			ctx := context.Background()
			wt, err := w.s.Writer(ctx)
			if err != nil {
				logr.FromContextOrDiscard(ctx).Error(err, "create writer failed")
				return
			}
			d := &storage.SizeWriter{}
			if err = jpeg.Encode(io.MultiWriter(wt, d), img, nil); err != nil {
				logr.FromContextOrDiscard(ctx).Error(err, "encoding failed")
				return
			}
			_ = wt.Commit(ctx, d.Size(), wt.Info().Digest(), blob.WithFromThough(model.TimeFromUnixNano(pkt.At.UnixNano())), blob.WithLabels(map[string][]string{
				"_device_id":  {pkt.ID},
				"_media_type": {mime.MediaTypeImageJPEG},
				"_size":       {strconv.Itoa(int(d.Size()))},
			}))
		}
	}
}

func (w *imageWriter) FrameFilename(pkt *livestream.Packet) string {
	buf := bytes.NewBuffer(nil)
	_ = w.t.Execute(buf, map[string]any{
		"Timestamp": pkt.At.Unix(),
	})
	return fmt.Sprintf("%s/%s", pkt.ID, buf.String())
}
