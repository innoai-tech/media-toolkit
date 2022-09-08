package rtsp

import (
	"fmt"
	"image"
	"io"
	"reflect"
	"unsafe"
)

// #cgo pkg-config: libavformat libavutil libavcodec
// #include <libavcodec/avcodec.h>
// #include <libavutil/avutil.h>
import "C"

func newDecoder(codecPar *C.AVCodecParameters) (dec *decoder, err error) {
	codec := C.avcodec_find_decoder(codecPar.codec_id)
	if codec == nil {
		return nil, fmt.Errorf("avcodec_find_decoder() failed")
	}

	dec = &decoder{}

	codecCtx := C.avcodec_alloc_context3(codec)
	if codecCtx == nil {
		return nil, fmt.Errorf("avcodec_alloc_context3() failed")
	}
	dec.codecCtx = codecCtx

	C.avcodec_parameters_to_context(dec.codecCtx, codecPar)
	if ret := C.avcodec_open2(dec.codecCtx, codec, nil); ret < 0 {
		return nil, fmt.Errorf("avcodec_open2() failed")
	}
	frame := C.av_frame_alloc()
	if frame == nil {
		return nil, fmt.Errorf("av_frame_alloc() failed")
	}
	dec.frame = frame

	return
}

type decoder struct {
	codecParameters *C.AVCodecParameters
	codecCtx        *C.AVCodecContext
	frame           *C.AVFrame
}

func (d *decoder) Close() error {
	d.codecParameters = nil

	if d.frame != nil {
		C.av_frame_free(&d.frame)
	}

	if d.codecCtx != nil {
		C.avcodec_close(d.codecCtx)
	}

	return nil
}

func (d *decoder) ToImage(pkt *C.AVPacket) (bool, *image.YCbCr, error) {
	if ret := C.avcodec_send_packet(d.codecCtx, pkt); ret < 0 {
		return false, nil, fmt.Errorf("avcodec_send_packet failed %d", int(ret))
	}

	ret := C.avcodec_receive_frame(d.codecCtx, d.frame)
	if ret == C.AVERROR_EOF {
		return false, nil, io.EOF
	}
	if ret == -35 {
		return false, nil, nil
	}

	w := int(d.frame.width)
	h := int(d.frame.height)
	ys := int(d.frame.linesize[0])
	cs := int(d.frame.linesize[1])

	return true, &image.YCbCr{
		Rect:           image.Rect(0, 0, w, h),
		Y:              fromCPtr(unsafe.Pointer(d.frame.data[0]), ys*h),
		Cb:             fromCPtr(unsafe.Pointer(d.frame.data[1]), cs*h/2),
		Cr:             fromCPtr(unsafe.Pointer(d.frame.data[2]), cs*h/2),
		YStride:        ys,
		CStride:        cs,
		SubsampleRatio: image.YCbCrSubsampleRatio420,
	}, nil
}

func fromCPtr(buf unsafe.Pointer, size int) (ret []uint8) {
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&ret))
	hdr.Cap = size
	hdr.Len = size
	hdr.Data = uintptr(buf)
	return
}
