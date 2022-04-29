package codec

import (
	"errors"
	"fmt"
	"image"
	"reflect"
	"unsafe"
)

/*
	#cgo CFLAGS: -I/usr/pkg/ffmpeg/include
	#cgo LDFLAGS: -L/usr/pkg/ffmpeg/lib -lavutil -lavcodec -lavformat
	#include <libavcodec/avcodec.h>
	#include <libavformat/avformat.h>
	#include <libavutil/avutil.h>

	typedef struct {
		const AVCodec *codec;
		AVCodecContext *ctx;
		AVFrame *frame;
		int got;
	} h264decoder_t;

	static int h264dec_new(h264decoder_t *dec, uint8_t *data, int len) {
		dec->codec = avcodec_find_decoder(AV_CODEC_ID_H264);
		dec->ctx = avcodec_alloc_context3(dec->codec);
		dec->frame = av_frame_alloc();
		dec->ctx->extradata = data;
		dec->ctx->extradata_size = len;
		dec->ctx->debug = 0x3;
		dec->ctx->time_base.num = 1001;
		dec->ctx->time_base.den = 30000;
		return avcodec_open2(dec->ctx, dec->codec, 0);
	}

	static int h264dec_close(h264decoder_t *dec) {
		av_frame_free(&dec->frame);
		avcodec_close(dec->ctx);
		return 0;
	}

	static int h264dec_decode(h264decoder_t *dec, uint8_t *data, int len) {
		AVPacket* pkt = av_packet_alloc();
		pkt->data = data;
		pkt->size = len;

		int response = 0;

		if (dec->ctx-> codec_type == AVMEDIA_TYPE_VIDEO || dec->ctx-> codec_type == AVMEDIA_TYPE_AUDIO) {
			response = avcodec_send_packet(dec->ctx, pkt);
			if (response < 0 && response != AVERROR(EAGAIN) && response != AVERROR_EOF) {
			} else {
				if (response >= 0) pkt->size = 0;
				response = avcodec_receive_frame(dec->ctx, dec->frame);
				if (response >= 0) dec->got = 1;
			}
		}
		return response;
	}

	static void libav_init() {
		avformat_network_init();
		// av_log_set_level(AV_LOG_DEBUG);
	}
*/
import "C"

type H264Decoder struct {
	dec C.h264decoder_t
}

func NewH264Decoder(header []byte) (dec *H264Decoder, err error) {
	dec = &H264Decoder{}
	r := C.h264dec_new(
		&dec.dec,
		(*C.uint8_t)(unsafe.Pointer(&header[0])),
		(C.int)(len(header)),
	)
	if int(r) < 0 {
		err = errors.New("open codec failed")
	}
	return
}

func (m *H264Decoder) Close() error {
	_ = C.h264dec_close(&m.dec)
	return nil
}

func (m *H264Decoder) Decode(nal []byte) (f *image.YCbCr, err error) {
	defer func() {
		if e := recover(); e != nil {
			if ee, ok := e.(error); ok {
				err = ee
			} else {
				err = fmt.Errorf("%v", e)
			}
		}
	}()
	r := C.h264dec_decode(
		&m.dec,
		(*C.uint8_t)(unsafe.Pointer(&nal[0])),
		(C.int)(len(nal)),
	)
	if int(r) < 0 {
		err = errors.New("decode failed")
		return
	}
	if m.dec.got == 0 {
		err = errors.New("no picture")
		return
	}

	w := int(m.dec.frame.width)
	h := int(m.dec.frame.height)
	ys := int(m.dec.frame.linesize[0])
	cs := int(m.dec.frame.linesize[1])

	f = &image.YCbCr{
		Y:              fromCPtr(unsafe.Pointer(m.dec.frame.data[0]), ys*h),
		Cb:             fromCPtr(unsafe.Pointer(m.dec.frame.data[1]), cs*h/2),
		Cr:             fromCPtr(unsafe.Pointer(m.dec.frame.data[2]), cs*h/2),
		YStride:        ys,
		CStride:        cs,
		SubsampleRatio: image.YCbCrSubsampleRatio420,
		Rect:           image.Rect(0, 0, w, h),
	}

	return
}

func init() {
	C.libav_init()
}

func fromCPtr(buf unsafe.Pointer, size int) (ret []uint8) {
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&ret))
	hdr.Cap = size
	hdr.Len = size
	hdr.Data = uintptr(buf)
	return
}
