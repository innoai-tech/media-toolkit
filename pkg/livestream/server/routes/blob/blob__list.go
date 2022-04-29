package blob

import (
	"context"
	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"github.com/innoai-tech/media-toolkit/pkg/types"
)

func init() {
	BlobRouter.Register(courier.NewRouter(&ListBlob{}))
}

type ListBlob struct {
	httpx.MethodGet `path:"/blobs"`
	TimeRange       types.DateTimeRange `name:"time" in:"query"`
	Filter          types.Filter        `name:"filter,omitempty" in:"query"`
}

func (req *ListBlob) Output(ctx context.Context) (any, error) {
	s := storage.StoreFromContext(ctx)

	return s.Query(
		ctx,
		blob.TimeRange{From: req.TimeRange.From, Through: req.TimeRange.To},
		blob.DefaultUser,
		req.Filter.Matchers...,
	)
}
