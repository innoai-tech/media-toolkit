package image

import (
	"context"
	"image/jpeg"
	"io"
	"strconv"
	"time"

	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/codec"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"github.com/innoai-tech/media-toolkit/pkg/storage/mime"
	"github.com/innoai-tech/media-toolkit/pkg/types"
	"github.com/innoai-tech/media-toolkit/pkg/util/syncutil"
	"github.com/pkg/errors"

	"github.com/deepch/vdk/codec/h264parser"
	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
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
		if cd, ok := c.(h264parser.CodecData); ok {
			if pkt.IsKeyFrame {
				go w.takePic(&cd, &pkt)
			}
		}
	}
}

func (w *imageWriter) takePic(codecData *h264parser.CodecData, pkt *livestream.Packet) (e error) {
	defer func() {
		if e != nil {
			w.l.Error(e, "take pic")
		}
		w.SendDone(e)
	}()

	decoder, err := codec.NewH264Decoder(codecData.Record)
	if err != nil {
		return err
	}
	defer decoder.Close()

	img, err := decoder.Decode(pkt.Data)
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
