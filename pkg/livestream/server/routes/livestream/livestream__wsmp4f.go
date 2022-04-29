package livestream

import (
	"context"
	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	"github.com/innoai-tech/media-toolkit/pkg/livestream/observer/wsmp4f"
	"net/http"
)

func init() {
	LiveStreamRouter.Register(courier.NewRouter(&LiveStreamWsmp4f{}))
}

type LiveStreamWsmp4f struct {
	httpx.MethodGet `path:"/live-streams/:id/wsmp4f"`
	ID              string `name:"id" in:"path"`
}

func (req *LiveStreamWsmp4f) Output(ctx context.Context) (any, error) {
	hub := livestream.StreamHubFromContext(ctx)
	return &upgrader{hub: hub, id: req.ID}, nil
}

type upgrader struct {
	hub *livestream.StreamHub
	id  string
}

func (ug *upgrader) Upgrade(rw http.ResponseWriter, req *http.Request) error {
	u := &websocket.Upgrader{
		ReadBufferSize: 1024, WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	ctx := req.Context()

	c, err := u.Upgrade(rw, req, nil)
	if err != nil {
		return err
	}

	sub, err := ug.hub.Subscribe(ug.id, wsmp4f.New(ctx, c))
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, "subscribe failed")
		return err
	}

	done := make(chan struct{})

	go func() {
		for {
			_, _, _ = c.NextReader()
			_ = c.Close()
			break
		}
	}()

	c.SetCloseHandler(func(code int, text string) error {
		close(done)
		return sub.Close()
	})

	<-done

	return nil
}
