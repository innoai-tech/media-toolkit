package codec

import "C"
import (
	"reflect"
	"unsafe"
)

/*
#cgo pkg-config: libavformat libavutil libavcodec
#include <libavutil/avutil.h>
#include <libavformat/avformat.h>
static void libav_init() {
	avformat_network_init();
	// av_log_set_level(AV_LOG_DEBUG);
}
*/
import "C"

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
