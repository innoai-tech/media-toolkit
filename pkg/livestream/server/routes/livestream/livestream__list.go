package livestream

import (
	"context"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
)

func init() {
	LiveStreamRouter.Register(courier.NewRouter(&ListLiveStream{}))
}

type ListLiveStream struct {
	httpx.MethodGet `path:"/live-streams"`
}

func (req *ListLiveStream) Output(ctx context.Context) (any, error) {
	hub := livestream.StreamHubFromContext(ctx)
	return hub.List(), nil
}
