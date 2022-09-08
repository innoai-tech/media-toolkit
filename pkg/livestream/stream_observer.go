package livestream

import (
	"github.com/pion/mediadevices"
	"io"
)

type StreamObserver interface {
	Name() string
	OnVideoSource(videoSource mediadevices.VideoSource)
	Done() <-chan error
	io.Closer
}

type CanUniqueKey interface {
	UniqueKey() any
}

func WithUniqueKey(key any, so StreamObserver) StreamObserver {
	return &streamObserverWithUniqKey{
		StreamObserver: so,
		key:            key,
	}
}

type streamObserverWithUniqKey struct {
	key any
	StreamObserver
}

func (s *streamObserverWithUniqKey) UniqueKey() any {
	return s.key
}
