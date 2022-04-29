package livestream

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/innoai-tech/media-toolkit/pkg/statuserr"

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
	hub.streams.Range(func(_, value any) bool {
		list = append(list, value.(StreamSubject).Info())
		return true
	})
	slices.SortFunc(list, func(a, b Stream) bool { return a.ID < b.ID })
	return list
}

func (hub *StreamHub) AddStream(s Stream) {
	hub.streams.Store(s.ID, NewStreamSubject(s))
}

func (hub *StreamHub) Status(ctx context.Context, id string) (Status, error) {
	s, ok := hub.streams.Load(id)
	if !ok {
		return Status{}, statuserr.Wrap(http.StatusNotFound, StreamNotFound, fmt.Sprintf("`%s` is not found", id))
	}
	return s.Status(), nil
}

func (hub *StreamHub) Subscribe(ctx context.Context, id string, ob StreamObserver) (io.Closer, error) {
	s, ok := hub.streams.Load(id)
	if !ok {
		return nil, statuserr.Wrap(http.StatusNotFound, StreamNotFound, fmt.Sprintf("`%s` is not found", id))
	}
	return s.Subscribe(ctx, ob)
}

type streamHubContextKey struct {
}

func StreamHubFromContext(ctx context.Context) *StreamHub {
	return ctx.Value(streamHubContextKey{}).(*StreamHub)
}

func NewContextWithStreamHub(ctx context.Context, hub *StreamHub) context.Context {
	return context.WithValue(ctx, streamHubContextKey{}, hub)
}
