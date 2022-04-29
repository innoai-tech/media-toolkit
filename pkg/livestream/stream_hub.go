package livestream

import (
	"context"
	"io"

	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

var (
	StreamNotFound = errors.New("stream not found")
)

type Stream struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Rtsp string `json:"rtsp"`
}

func NewStreamHub() *StreamHub {
	return &StreamHub{}
}

type StreamHub struct {
	streams XMap[string, StreamSubject]
}

func (hub *StreamHub) List() []Stream {
	list := make([]Stream, 0)
	hub.streams.Range(func(key, value any) bool {
		list = append(list, value.(StreamSubject).Info())
		return true
	})
	slices.SortFunc(list, func(a, b Stream) bool { return a.ID < b.ID })
	return list
}

func (hub *StreamHub) AddStream(s Stream) {
	hub.streams.Store(s.ID, NewStreamSubject(s))
}

func (hub *StreamHub) Subscribe(id string, ob StreamObserver) (io.Closer, error) {
	s, ok := hub.streams.Load(id)
	if !ok {
		return nil, StreamNotFound
	}
	return s.Subscribe(ob)
}

type streamHubContextKey struct {
}

func StreamHubFromContext(ctx context.Context) *StreamHub {
	return ctx.Value(streamHubContextKey{}).(*StreamHub)
}

func NewContextWithStreamHub(ctx context.Context, hub *StreamHub) context.Context {
	return context.WithValue(ctx, streamHubContextKey{}, hub)
}
