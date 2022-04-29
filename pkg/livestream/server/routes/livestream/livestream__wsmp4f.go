package livestream

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	"github.com/innoai-tech/media-toolkit/pkg/livestream/observer/wsmp4f"
	"github.com/innoai-tech/media-toolkit/pkg/util/syncutil"
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
	id  string
	hub *livestream.StreamHub
}

func (ug *upgrader) Upgrade(rw http.ResponseWriter, req *http.Request) error {
	u := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
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

	sub, err := ug.hub.Subscribe(ctx, ug.id, wsmp4f.New(ctx, &wsWriter{c: c}))
	if err != nil {
		logr.FromContextOrDiscard(ctx).Error(err, "subscribe failed")
		return err
	}
	defer sub.Close()

	done := syncutil.NewChan[struct{}]()

	go func() {
		// just push live stream,
		// any msg from client should close
		_, _, _ = c.NextReader()
		done.Close()
	}()

	go func() {
		<-ctx.Done()
		done.Close()
	}()

	<-done.Recv()

	return nil
}

type wsWriter struct {
	c *websocket.Conn
}

func (w *wsWriter) Write(p []byte) (n int, err error) {
	return len(p), w.c.WriteMessage(websocket.BinaryMessage, p)
}
