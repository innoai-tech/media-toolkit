package livestream

import (
	"context"
	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	"github.com/innoai-tech/media-toolkit/pkg/livestream/observer/image"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
)

func init() {
	LiveStreamRouter.Register(courier.NewRouter(&LiveStreamTakePic{}))
}

type LiveStreamTakePic struct {
	httpx.MethodPut `path:"/live-streams/:id/takepic"`
	ID              string `name:"id" in:"path"`
}

func (req *LiveStreamTakePic) Output(ctx context.Context) (any, error) {
	hub := livestream.StreamHubFromContext(ctx)
	store := storage.StoreFromContext(ctx)

	sub, err := hub.Subscribe(req.ID, image.New(store, image.Options{Filename: "{{ .Timestamp }}"}))
	if err != nil {
		return nil, err
	}
	// todo wait
	_ = sub.Close()
	return nil, nil
}
