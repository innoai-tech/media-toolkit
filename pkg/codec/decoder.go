package codec

import (
	"github.com/pkg/errors"
	"image"
	"reflect"
	"unsafe"
)

var EAGAIN = errors.New("EAGAIN")

type Decoder interface {
	Info() Info
	DecodeToImage(nal []byte) (f *image.YCbCr, err error)
	Close() error
}

func annexbNALUStartCode() []byte {
	return []byte{0x00, 0x00, 0x00, 0x01}
}

func fromCPtr(buf unsafe.Pointer, size int) (ret []uint8) {
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&ret))
	hdr.Cap = size
	hdr.Len = size
	hdr.Data = uintptr(buf)
	return
}

type Info struct {
	Width  int
	Height int
}
