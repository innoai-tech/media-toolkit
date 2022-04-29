package livestream

import (
	"context"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"io"
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
	streams  XMap[string, Stream]
	connects XMap[string, StreamConn]
}

func (hub *StreamHub) List() []Stream {
	list := make([]Stream, 0)
	hub.streams.Range(func(key, value any) bool {
		list = append(list, value.(Stream))
		return true
	})
	slices.SortFunc(list, func(a, b Stream) bool { return a.ID < b.ID })
	return list
}

func (hub *StreamHub) AddStream(s Stream) {
	hub.streams.Store(s.ID, s)
}

func (hub *StreamHub) ConnStream(ctx context.Context, id string, streamWorker StreamWriter) (io.Closer, error) {
	s, ok := hub.streams.Load(id)
	if !ok {
		return nil, StreamNotFound
	}

	c, ok := hub.connects.Load(id)
	if !ok || c.IsClosed() {
		c = NewStreamConn(ctx, id)

		if err := c.Dial(s.Rtsp); err != nil {
			return nil, err
		}

		go func() {
			_ = c.Serve(streamWorker)
		}()

		hub.connects.Store(id, c)
	}

	return c, nil
}
