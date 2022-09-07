package image

import "image"

type Decoder interface {
	DecodeToImage(nal []byte) (f *image.YCbCr, err error)
	Close() error
}
