package codec

import (
	"fmt"
	"image"
	"unsafe"
)

// #include <libavcodec/avcodec.h>
// #include <libavutil/avutil.h>
import "C"

type H265Decoder struct {
	codecCtx *C.AVCodecContext
	frame    *C.AVFrame
	width    int
	height   int
}

func NewH265Decoder(avcDecoderConfRecord []byte, width, height int) (*H265Decoder, error) {
	if len(avcDecoderConfRecord) == 0 {
		return nil, fmt.Errorf("h265 decoder need AVCDecoderConfRecord")
	}

	codec := C.avcodec_find_decoder(C.AV_CODEC_ID_H265)
	if codec == nil {
		return nil, fmt.Errorf("avcodec_find_decoder() failed")
	}

	codecCtx := C.avcodec_alloc_context3(codec)
	if codecCtx == nil {
		return nil, fmt.Errorf("avcodec_alloc_context3() failed")
	}

	codecCtx.extradata_size = (C.int)(len(avcDecoderConfRecord))
	codecCtx.extradata = (*C.uint8_t)(unsafe.Pointer(&avcDecoderConfRecord[0]))

	res := C.avcodec_open2(codecCtx, codec, nil)
	if res < 0 {
		C.avcodec_close(codecCtx)
		return nil, fmt.Errorf("avcodec_open2() failed")
	}

	frame := C.av_frame_alloc()
	if frame == nil {
		C.avcodec_close(codecCtx)
		return nil, fmt.Errorf("av_frame_alloc() failed")
	}

	return &H265Decoder{codecCtx: codecCtx, frame: frame, width: width, height: height}, nil
}

func (d *H265Decoder) Info() Info {
	return Info{Width: d.width, Height: d.height}
}

func (d *H265Decoder) Close() error {
	if d.frame != nil {
		C.av_frame_free(&d.frame)
	}
	C.avcodec_close(d.codecCtx)
	return nil
}

func (d *H265Decoder) DecodeToImage(nalu []byte) (*image.YCbCr, error) {
	avPacket := C.AVPacket{}
	avPacket.size = C.int(len(nalu))
	avPacket.data = (*C.uint8_t)(C.CBytes(nalu))
	defer C.free(unsafe.Pointer(avPacket.data))

	res := C.avcodec_send_packet(d.codecCtx, &avPacket)
	if res < 0 {
		return nil, fmt.Errorf("avcodec_send_packet() failed with %d", int(res))
	}

	for {
		res = C.avcodec_receive_frame(d.codecCtx, d.frame)
		if res == C.AVERROR_EOF {
			break
		}
		if res == -35 {
			return nil, EAGAIN
		}
		if res >= 0 {
			break
		}
	}

	if d.frame == nil {
		return nil, fmt.Errorf("avcodec_receive_frame() failed with %d", int(res))
	}

	w := int(d.frame.width)
	h := int(d.frame.height)
	ys := int(d.frame.linesize[0])
	cs := int(d.frame.linesize[1])

	return &image.YCbCr{
		Rect:           image.Rect(0, 0, w, h),
		Y:              fromCPtr(unsafe.Pointer(d.frame.data[0]), ys*h),
		Cb:             fromCPtr(unsafe.Pointer(d.frame.data[1]), cs*h/2),
		Cr:             fromCPtr(unsafe.Pointer(d.frame.data[2]), cs*h/2),
		YStride:        ys,
		CStride:        cs,
		SubsampleRatio: image.YCbCrSubsampleRatio420,
	}, nil
}
