package livestream

import (
	"context"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
)

func init() {
	LiveStreamRouter.Register(courier.NewRouter(&LiveStreamStatus{}))
}

type LiveStreamStatus struct {
	httpx.MethodGet `path:"/live-streams/:id/status"`
	ID              string `name:"id" in:"path"`
}

func (req *LiveStreamStatus) Output(ctx context.Context) (any, error) {
	hub := livestream.StreamHubFromContext(ctx)
	return hub.Status(ctx, req.ID)
}
