package livestream

import (
	"context"
	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	"github.com/innoai-tech/media-toolkit/pkg/livestream/observer/wsmp4f"
	"github.com/innoai-tech/media-toolkit/pkg/util/syncutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	ctx, cancel := signal.NotifyContext(req.Context(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	c, err := u.Upgrade(rw, req.WithContext(ctx), nil)
	if err != nil {
		return err
	}
	defer c.Close()

	sub, err := ug.hub.Subscribe(ctx, ug.id, wsmp4f.New(ctx, c))
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, "subscribe failed")
		return err
	}
	defer sub.Close()

	done := syncutil.NewChan[struct{}]()

	go func() {
		// just push live stream,
		// any msg from client should close
		for {
			_, _, _ = c.NextReader()
			done.Close()
			break
		}
	}()

	go func() {
		<-ctx.Done()
		done.Close()
	}()

	c.SetCloseHandler(func(code int, text string) error {
		done.Close()
		return nil
	})

	<-done.Recv()

	return nil
}
