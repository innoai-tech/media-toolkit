package livestream

import (
	"context"
	"time"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	"github.com/innoai-tech/media-toolkit/pkg/livestream/observer/video"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
)

func init() {
	LiveStreamRouter.Register(courier.NewRouter(&LiveStreamTakeVideo{}))
}

type LiveStreamTakeVideo struct {
	httpx.MethodPut `path:"/live-streams/:id/takevideo"`
	ID              string `name:"id" in:"path"`
	Stop            bool   `name:"stop,omitempty" in:"query"`
}

func (req *LiveStreamTakeVideo) Output(ctx context.Context) (any, error) {
	hub := livestream.StreamHubFromContext(ctx)
	store := storage.StoreFromContext(ctx)

	s, err := hub.Subscribe(ctx, req.ID, livestream.WithUniqueKey(req.ID, video.New(store, func(o *video.Options) {
		o.MaxDuration = 60 * 10 * time.Second
	})))
	if err != nil {
		return nil, err
	}

	if req.Stop {
		return nil, s.Close()
	}

	return nil, nil
}
