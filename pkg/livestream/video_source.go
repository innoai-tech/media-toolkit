package livestream

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/livestream/core"
	"github.com/innoai-tech/media-toolkit/pkg/mediadevice/rtsp"
	"github.com/innoai-tech/media-toolkit/pkg/util/syncutil"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec"
	"github.com/pion/mediadevices/pkg/codec/x264"
	"github.com/pion/mediadevices/pkg/io"
	"github.com/pion/mediadevices/pkg/io/video"
	"image"
	"sync/atomic"
	"time"
)

type EncodingPreset string

const (
	Preset1080P EncodingPreset = "1080P"
	Preset720P  EncodingPreset = "720P"
	Preset480P  EncodingPreset = "480P"
)

type VideoSource interface {
	ID() string
	NewReader() (video.Reader, error)
	NewEncodedReader(EncodingPreset) (mediadevices.EncodedReadCloser, error)
	Status() Status
}

type VideoTrack interface {
	NewEncodedReader() (mediadevices.EncodedReadCloser, error)
}

func newVideoSource(ctx context.Context, stream core.Stream, getStatus func() Status) VideoSource {
	vs := &videoSource{
		l:           logr.FromContextOrDiscard(ctx).WithValues("stream_id", stream.ID, "stream_name", stream.Name),
		stream:      stream,
		getStatus:   getStatus,
		idleTimeout: 30 * time.Second,
	}

	vs.videoSource = syncutil.NewPool(func() (mediadevices.VideoSource, error) {
		vs.l.Info(fmt.Sprintf("%s starting...", vs.stream.Name))

		ctx, cancel := context.WithTimeout(context.Background(), vs.idleTimeout)
		defer cancel()

		videoSrc, err := rtsp.Open(ctx, vs.stream.Rtsp, vs.stream.ID)
		if err != nil {
			return nil, err
		}

		vs.idleTimer = time.NewTimer(vs.idleTimeout)

		go func() {
			// cleanup and wait next initial.
			defer vs.videoSource.Put(nil)

			select {
			case <-vs.idleTimer.C:
				vs.l.Info(fmt.Sprintf("auto closed when idle %s", vs.idleTimeout))
				_ = videoSrc.Close()
			}
		}()

		return videoSrc, nil
	})

	for _, preset := range []EncodingPreset{Preset1080P, Preset720P, Preset480P} {
		vt := func(vs mediadevices.VideoSource, preset EncodingPreset) (p *syncutil.Pool[VideoTrack]) {
			return syncutil.NewPool(func() (VideoTrack, error) {
				x264Params, _ := x264.NewParams()
				x264Params.Preset = x264.PresetMedium
				x264Params.BitRate = 4_000_000

				codecSelector := mediadevices.NewCodecSelector(
					mediadevices.WithVideoEncoders(&x264Params),
				)

				videoTrack := mediadevices.NewVideoTrack(vs, codecSelector)

				encodedReadCloser, err := videoTrack.NewEncodedReader(x264Params.RTPCodec().MimeType)
				if err != nil {
					return nil, err
				}

				broadcaster := io.NewBroadcaster(io.ReaderFunc(func() (interface{}, func(), error) {
					return encodedReadCloser.Read()
				}), nil)

				return &videoEncodedBroadcaster{
					closeFn: func() error {
						p.Put(nil)
						broadcaster = nil
						codecSelector = nil
						_ = videoTrack.Close()
						return encodedReadCloser.Close()
					},
					broadcaster: broadcaster,
				}, nil
			})
		}(vs, preset)

		vs.videoTracks.Store(preset, vt)
	}

	vs.broadcaster = video.NewBroadcaster(vs, nil)
	return vs
}

type videoSource struct {
	l logr.Logger

	getStatus func() Status

	stream core.Stream

	idleTimeout time.Duration
	idleTimer   *time.Timer

	videoSource *syncutil.Pool[mediadevices.VideoSource]
	videoTracks syncutil.Map[EncodingPreset, *syncutil.Pool[VideoTrack]]
	broadcaster *video.Broadcaster
}

func (s *videoSource) Close() error {
	return nil
}

func (s *videoSource) Status() Status {
	return s.getStatus()
}

func (s *videoSource) ID() string {
	return s.stream.ID
}

func (s *videoSource) Read() (image.Image, func(), error) {
	vs, err := s.videoSource.Get()
	if err != nil {
		return nil, nil, err
	}
	s.idleTimer.Reset(s.idleTimeout)
	return vs.Read()
}

func (s *videoSource) NewReader() (video.Reader, error) {
	return s.broadcaster.NewReader(true), nil
}

func (s *videoSource) NewEncodedReader(preset EncodingPreset) (mediadevices.EncodedReadCloser, error) {
	vt, ok := s.videoTracks.Load(preset)
	if !ok {
		return nil, fmt.Errorf("unsuportted preset %s", preset)
	}

	v, err := vt.Get()
	if err != nil {
		return nil, err
	}

	return v.NewEncodedReader()
}

type videoEncodedBroadcaster struct {
	broadcaster *io.Broadcaster
	used        int64
	closeFn     func() error
}

func (b *videoEncodedBroadcaster) NewEncodedReader() (mediadevices.EncodedReadCloser, error) {
	atomic.AddInt64(&b.used, 1)

	return &videoEncodedReadCloser{
		closeFn: func() error {
			atomic.AddInt64(&b.used, -1)
			if atomic.LoadInt64(&b.used) == 0 {
				return b.closeFn()
			}
			return nil
		},
		broadcasterReader: b.broadcaster.NewReader(func(v interface{}) interface{} {
			eb := v.(mediadevices.EncodedBuffer)
			return mediadevices.EncodedBuffer{
				Data:    eb.Data[0:],
				Samples: eb.Samples,
			}
		}),
	}, nil
}

type videoEncodedReadCloser struct {
	closeFn           func() error
	broadcasterReader io.Reader
}

func (v *videoEncodedReadCloser) Read() (mediadevices.EncodedBuffer, func(), error) {
	buf, release, err := v.broadcasterReader.Read()
	return buf.(mediadevices.EncodedBuffer), release, err
}

func (v *videoEncodedReadCloser) Close() error {
	return v.closeFn()
}

func (v *videoEncodedReadCloser) Controller() codec.EncoderController {
	return nil
}
