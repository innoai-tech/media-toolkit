package livestream

import (
	"io"
)

type StreamNotifier interface {
	WritePacket(pkt Packet)
}

type StreamObserver interface {
	StreamNotifier
	Name() string
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
