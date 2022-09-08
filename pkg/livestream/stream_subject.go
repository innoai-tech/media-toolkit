package livestream

import (
	"context"
	"fmt"
	"github.com/innoai-tech/media-toolkit/pkg/livestream/core"
	"github.com/innoai-tech/media-toolkit/pkg/util/syncutil"
	"io"
	"time"

	"github.com/go-logr/logr"
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
	At        time.Time      `json:"at"`
	Observers map[string]int `json:"observers"`
}

func NewStreamSubject(ctx context.Context, stream core.Stream) StreamSubject {
	ss := &streamSubject{
		stream: stream,
	}

	ss.videoSource = syncutil.NewPool(func() (VideoSource, error) {
		return newVideoSource(ctx, stream, ss.Status), nil
	})

	return ss
}

type streamSubject struct {
	stream      core.Stream
	videoSource *syncutil.Pool[VideoSource]
	observers   syncutil.Map[any, StreamObserver]
}

func (s *streamSubject) Close() error {
	return nil
}

func (s *streamSubject) Name() string {
	return "StreamSubject"
}

func (s *streamSubject) Subscribe(ctx context.Context, o StreamObserver) (io.Closer, error) {
	l := logr.FromContextOrDiscard(ctx).
		WithValues(
			"stream_id", s.stream.ID,
			"stream_name", s.stream.Name,
		)

	ctx = logr.NewContext(ctx, l)

	videoSrc, err := s.videoSource.Get()
	if err != nil {
		return nil, err
	}

	var key any = o

	if can, ok := o.(CanUniqueKey); ok {
		key = fmt.Sprintf("%s:%s", o.Name(), can.UniqueKey())
	}

	if found, ok := s.observers.Load(key); ok {
		return found, nil
	}

	go func(videoSrc VideoSource) {
		defer s.observers.Delete(key)
		defer o.Close()

		o.OnVideoSource(ctx, videoSrc)
		<-o.Done()

		l.V(1).
			WithValues("status", videoSrc.Status()).
			Info(fmt.Sprintf("stream observer `%s` removed", o.Name()))
	}(videoSrc)

	s.observers.Store(key, o)

	l.V(1).
		WithValues("status", videoSrc.Status()).
		Info(fmt.Sprintf("stream observer `%s` added.", o.Name()))

	return o, nil
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

	status.Active = s.videoSource != nil

	return status
}

func (s *streamSubject) Info() core.Stream {
	return s.stream
}
