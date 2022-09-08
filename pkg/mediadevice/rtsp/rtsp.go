package rtsp

import (
	"context"
	"fmt"
	"github.com/pion/mediadevices"
	"image"
	"io"
	"sync"
	"sync/atomic"
	"unsafe"
	_ "unsafe"
)

// #cgo pkg-config: libavformat libavutil libavcodec
// #include <libavcodec/avcodec.h>
// #include <libavformat/avformat.h>
// #include <libavutil/avutil.h>
// void * ptr_at(void **ptr, int idx) {
//    return ptr[idx];
// }
import "C"

func init() {
	C.avformat_network_init()
}

func Open(ctx context.Context, rtsp string, id string) (conn *RTSPConnect, err error) {
	select {
	case <-ctx.Done():
		err = ctx.Err()
	default:
		formatCtx := C.avformat_alloc_context()
		if formatCtx == nil {
			return nil, fmt.Errorf("avformat_alloc_context() failed")
		}

		opts := (*C.AVDictionary)(nil)
		C.av_dict_set(&opts, C.CString("rtsp_transport"), C.CString("tcp"), 0)

		if ret := C.avformat_open_input(&formatCtx, C.CString(rtsp), nil, &opts); int(ret) < 0 {
			return nil, fmt.Errorf("avformat_open_input() failed %d", int(ret))
		}

		if ret := C.avformat_find_stream_info(formatCtx, nil); int(ret) < 0 {
			return nil, fmt.Errorf("avformat_find_stream_info() failed %d", int(ret))
		}

		conn = &RTSPConnect{id: id, formatCtx: formatCtx}
		defer func() {
			if err != nil {
				_ = conn.Close()
			}
		}()

		for i := 0; i < int(formatCtx.nb_streams); i++ {
			if stream := streamAt(formatCtx, i); stream.codecpar.codec_type == C.AVMEDIA_TYPE_VIDEO {
				d, err := newDecoder(stream.codecpar)
				if err != nil {
					return nil, err
				}
				conn.videoStreamIndex = i
				conn.videoStreamDecoder = d
				break
			}
		}
	}

	return conn, nil
}

func streamAt(fctx *C.AVFormatContext, i int) *C.AVStream {
	p := (*unsafe.Pointer)(unsafe.Pointer(fctx.streams))
	return (*C.AVStream)(C.ptr_at(p, C.int(i)))
}

// TODO added audio source
var _ interface {
	mediadevices.VideoSource
} = &RTSPConnect{}

type RTSPConnect struct {
	id        string
	formatCtx *C.AVFormatContext

	videoStreamIndex   int
	videoStreamDecoder *decoder

	play sync.Once

	done int64

	wg sync.WaitGroup
}

func (c *RTSPConnect) ID() string {
	return c.id
}

func (c *RTSPConnect) Close() error {
	atomic.StoreInt64(&c.done, 1)
	// Must wait last Read finish
	c.wg.Wait()

	if videoStreamDecoder := c.videoStreamDecoder; videoStreamDecoder != nil {
		c.videoStreamDecoder = nil
		_ = videoStreamDecoder.Close()
	}

	if c.formatCtx != nil {
		C.avformat_free_context(c.formatCtx)
		c.formatCtx = nil
	}

	return nil
}

func (c *RTSPConnect) Read() (i image.Image, release func(), e error) {
	if atomic.LoadInt64(&c.done) != 0 {
		return nil, nil, io.EOF
	}

	c.wg.Add(1)
	defer c.wg.Done()

	for {
		ready, img, err := c.readImage()
		if err != nil {
			return nil, func() {}, err
		}
		if ready {
			return img, func() {}, nil
		}
	}
}

func (c *RTSPConnect) readImage() (bool, *image.YCbCr, error) {
	if c.videoStreamDecoder == nil {
		return false, nil, fmt.Errorf("decoder no init")
	}

	c.play.Do(func() {
		C.av_read_play(c.formatCtx)
	})

	pkt := C.av_packet_alloc()
	defer C.av_packet_free(&pkt)

	ret := C.av_read_frame(c.formatCtx, pkt)
	if int(ret) < 0 {
		return false, nil, fmt.Errorf("av_read_frame failed %d", int(ret))
	}

	// drop other streams
	if int(pkt.stream_index) != c.videoStreamIndex {
		return false, nil, nil
	}

	return c.videoStreamDecoder.ToImage(pkt)
}
