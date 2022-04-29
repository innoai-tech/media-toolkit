package image

import (
	"bytes"
	"fmt"
	"github.com/containerd/containerd/content"
	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/codec"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"image/jpeg"
	"io"
	"strconv"
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

	return &imageWriter{
		options: options,
		s:       s,
		t:       t,
	}
}

type imageWriter struct {
	s       storage.Store
	t       *template.Template
	options Options
	decoder *codec.H264Decoder
	livestream.Switch
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

func (w *imageWriter) WritePacket(pkt *livestream.Packet) {
	if w.decoder != nil {
		img, err := w.decoder.Decode(pkt.Data)
		if err == nil && img != nil {
			ctx := pkt.Context()
			ref := w.FrameFilename(pkt)
			wt, err := w.s.Writer(ctx, content.WithRef(ref))
			if err != nil {
				logr.FromContextOrDiscard(ctx).Error(err, "create writer failed")
				return
			}
			d := &storage.SizeWriter{}
			if err = jpeg.Encode(io.MultiWriter(wt, d), img, nil); err != nil {
				logr.FromContextOrDiscard(ctx).Error(err, "encoding failed")
				return
			}
			_ = wt.Commit(ctx, d.Size(), wt.Digest(), content.WithLabels(map[string]string{
				"$ref":      ref,
				"mediaType": storage.MediaTypeImageJPEG,
				"size":      strconv.Itoa(int(d.Size())),
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
