package blob

import (
	"context"
	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
)

func init() {
	BlobRouter.Register(courier.NewRouter(&DeleteBlob{}))
}

type DeleteBlob struct {
	httpx.MethodDelete `path:"/blobs/:ref"`
	Ref                blob.RefString `name:"ref" in:"path"`
}

func (req *DeleteBlob) Output(ctx context.Context) (any, error) {
	s := storage.StoreFromContext(ctx)
	return nil, s.Delete(ctx, req.Ref.Ref())
}
