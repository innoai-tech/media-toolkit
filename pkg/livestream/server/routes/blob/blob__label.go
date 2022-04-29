package blob

import (
	"context"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
)

func init() {
	BlobRouter.Register(courier.NewRouter(&LabelBlob{}))
}

type LabelBlob struct {
	httpx.MethodPut `path:"/blobs/:ref/labels/:name/:value"`
	Ref             blob.RefString `name:"ref" in:"path"`
	Name            string         `name:"name" in:"path"`
	Value           string         `name:"value" in:"path"`
}

func (req *LabelBlob) Output(ctx context.Context) (any, error) {
	s := storage.StoreFromContext(ctx)
	return nil, s.PutLabel(ctx, req.Ref.Ref(), req.Name, req.Value)
}
