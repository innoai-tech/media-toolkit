package livestream

import (
	"context"
	"fmt"
	"github.com/innoai-tech/media-toolkit/pkg/livestream/core"
	"github.com/innoai-tech/media-toolkit/pkg/mediadevice/rtsp"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/io/video"
	"image"
	"io"
	"sync"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/go-logr/logr"
	"github.com/innoai-tech/media-toolkit/pkg/util/syncutil"
	"github.com/pkg/errors"
)

var (
	StreamExitNoVideoOnStream = errors.New("stream exit no video on stream")
	StreamExitCodecChanged    = errors.New("stream exit when codec changed")
	StreamExitRtspDisconnect  = errors.New("stream exit rtsp disconnect")
	StreamExitIdleTimeout     = errors.New("stream exit idle timeout")
)

type StreamSubject interface {
	Info() core.Stream
	Status() Status
	Subscribe(ctx context.Context, o StreamObserver) (io.Closer, error)
}

type Status struct {
	Active    bool           `json:"active"`
	Observers map[string]int `json:"observers"`
}

type Metadata struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	At        time.Time      `json:"at"`
	Observers map[string]int `json:"observers"`
}

type Packet struct {
	Codecs []av.CodecData `json:"codecs,omitempty"`
	av.Packet
	Frame image.Image
	Metadata
}

func (p Packet) Clone() Packet {
	data := make([]byte, len(p.Data))
	copy(data, p.Data)
	p.Data = data
	return p
}

func NewStreamSubject(stream core.Stream) StreamSubject {
	return &streamSubject{
		stream: stream,
	}
}

type streamSubject struct {
	stream    core.Stream
	worker    syncutil.ValueMutex[*videoSource]
	observers sync.Map
}

func (s *streamSubject) Close() error {
	if w := s.worker.Get(); w != nil {
		defer func() {
			s.worker.Set(nil)
		}()
		return w.Close()
	}
	return nil
}

func (s *streamSubject) Name() string {
	return "StreamSubject"
}

func (s *streamSubject) Subscribe(ctx context.Context, o StreamObserver) (io.Closer, error) {
	if c, ok := o.(CanUniqueKey); ok {
		if ss, ok := s.observers.Load(c.UniqueKey()); ok {
			return ss.(StreamObserver), nil
		}
	}

	l := logr.FromContextOrDiscard(ctx).WithValues("stream_id", s.stream.ID, "stream_name", s.stream.Name)

	// try to serve rtspWorker when observer add
	if w := s.worker.Get(); w == nil {
		w := s.worker.Set(newVideoSource(logr.NewContext(ctx, l), s.stream))
		go func() {
			defer w.Close()
			err := <-w.Done()
			if err != nil {
				l.Error(err, "video source exit error")
			}
			s.worker.Set(nil)
		}()
	}

	ss := &streamObserverWrapper{
		StreamObserver: o,
		streamSubject:  s,
	}

	go func() {
		ss.OnVideoSource(s.worker.Get())

		for {
			select {
			case _ = <-ss.Done():
				l.V(1).Info(fmt.Sprintf("stream observer `%s` removed", o.Name()))
				_ = ss.Close()
				return
			}
		}
	}()

	l.V(1).Info(fmt.Sprintf("stream observer `%s` added.", o.Name()))
	s.observers.Store(ss.UniqueKey(), ss)

	return ss, nil
}

type streamObserverWrapper struct {
	StreamObserver
	streamSubject *streamSubject
}

func (s *streamObserverWrapper) UniqueKey() any {
	if can, ok := s.StreamObserver.(CanUniqueKey); ok {
		return can.UniqueKey()
	}
	return s
}

func (s *streamObserverWrapper) Close() error {
	defer func() {
		s.streamSubject.observers.Delete(s.UniqueKey())
	}()
	return s.StreamObserver.Close()
}

func (s *streamSubject) HasObserver() bool {
	observerCount := 0
	s.observers.Range(func(_, value any) bool {
		observerCount++
		return true
	})
	return observerCount > 0
}

func (s *streamSubject) Status() Status {
	status := Status{
		Observers: map[string]int{},
	}

	s.observers.Range(func(_, value any) bool {
		so := value.(StreamObserver)
		status.Observers[so.Name()] += 1
		return true
	})

	status.Active = s.worker.Get() != nil

	return status
}

func (s *streamSubject) Info() core.Stream {
	return s.stream
}

func newVideoSource(ctx context.Context, stream core.Stream) *videoSource {
	return &videoSource{
		CloseNotifier: syncutil.NewCloseNotifier(),
		l:             logr.FromContextOrDiscard(ctx),
		stream:        stream,
		idleTimeout:   30 * time.Second,
	}
}

type videoSource struct {
	l      logr.Logger
	stream core.Stream
	syncutil.CloseNotifier
	r           video.Reader
	videoSource mediadevices.VideoSource
	idleTimeout time.Duration
	idleTimer   *time.Timer
}

func (s *videoSource) Close() error {
	if !s.CloseNotifier.Closed() {
		defer func() {
			_ = s.CloseNotifier.Close()
		}()
		s.l.Info("shutting down...")
		return s.videoSource.Close()
	}
	return nil
}

func (s *videoSource) ID() string {
	return s.stream.ID
}

func (s *videoSource) Read() (img image.Image, release func(), err error) {
	if s.videoSource == nil {
		s.l.Info("starting...")
		ctx, _ := context.WithTimeout(context.Background(), s.idleTimeout)
		vs, err := rtsp.Open(ctx, s.stream.Rtsp, s.stream.ID)
		if err != nil {
			return nil, func() {}, err
		}
		s.videoSource = vs
		s.r = video.NewBroadcaster(vs, nil).NewReader(true)
		s.idleTimer = time.NewTimer(s.idleTimeout)
	}

	s.idleTimer.Reset(s.idleTimeout)
	return s.r.Read()
}
