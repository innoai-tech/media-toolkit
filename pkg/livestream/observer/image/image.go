package image

import (
	"context"
	"github.com/deepch/vdk/codec/h265parser"
	"image/jpeg"
	"io"
	"strconv"
	"time"

	"github.com/deepch/vdk/codec/h264parser"
	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	imagecodec "github.com/innoai-tech/media-toolkit/pkg/codec/image"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"github.com/innoai-tech/media-toolkit/pkg/storage/mime"
	"github.com/innoai-tech/media-toolkit/pkg/types"
	"github.com/innoai-tech/media-toolkit/pkg/util/syncutil"
	"github.com/pkg/errors"
)

type Options struct {
	Timeout time.Duration
}

type OptFunc = func(o *Options)

func New(ctx context.Context, s storage.Ingester, opts ...OptFunc) livestream.StreamObserver {
	options := &Options{
		Timeout: 10 * time.Second,
	}

	for i := range opts {
		opts[i](options)
	}

	return &imageWriter{
		options:       *options,
		s:             s,
		l:             logr.FromContextOrDiscard(ctx),
		CloseNotifier: syncutil.NewCloseNotifier(),
	}
}

type imageWriter struct {
	l       logr.Logger
	s       storage.Ingester
	options Options
	syncutil.CloseNotifier
}

func (w *imageWriter) Name() string {
	return "Image"
}

func (w *imageWriter) WritePacket(pkt livestream.Packet) {
	for _, c := range pkt.Codecs {
		// FIXME cache key frames and merge them
		if pkt.IsKeyFrame {
			switch x := c.(type) {
			case h264parser.CodecData:
				go w.takePic(func() (imagecodec.Decoder, error) {
					return imagecodec.NewH264Decoder(x.Record)
				}, &pkt)
			case h265parser.CodecData:
				go w.takePic(func() (imagecodec.Decoder, error) {
					return imagecodec.NewH265Decoder(x.Record)
				}, &pkt)
			}
		}
	}
}

func (w *imageWriter) takePic(newDecoder func() (imagecodec.Decoder, error), pkt *livestream.Packet) (e error) {
	defer func() {
		if e != nil {
			w.l.Error(e, "take pic")
		}
		w.SendDone(e)
	}()

	decoder, err := newDecoder()
	if err != nil {
		return err
	}
	defer decoder.Close()

	img, err := decoder.DecodeToImage(pkt.Data)
	if err != nil {
		return errors.Wrap(err, "decode failed")
	}

	ctx := context.Background()

	wt, err := w.s.Writer(ctx)
	if err != nil {
		return errors.Wrap(err, "create writer failed")
	}

	d := &storage.SizeWriter{}
	if err = jpeg.Encode(io.MultiWriter(wt, d), img, nil); err != nil {
		return errors.Wrap(err, "jpeg encoding failed")
	}

	err = wt.Commit(ctx, d.Size(), wt.Info().Digest(), blob.WithFromThough(types.TimeFromUnixNano(pkt.At.UnixNano())), blob.WithLabels(map[string][]string{
		"_device_id":  {pkt.ID},
		"_media_type": {mime.MediaTypeImageJPEG},
		"_size":       {strconv.Itoa(int(d.Size()))},
	}))
	if err != nil {
		return errors.Wrap(err, "commit failed")
	}
	return nil
}
