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
	BlobRouter.Register(courier.NewRouter(&ExportDataset{}))
}

type ExportDataset struct {
	httpx.MethodGet `path:"/datasets"`
	TimeRange       types.DateTimeRange `name:"time" in:"query"`
	Filter          types.Filter        `name:"filter,omitempty" in:"query"`
}

func (req *ExportDataset) Output(ctx context.Context) (any, error) {
	s := storage.StoreFromContext(ctx)
	blobs, err := s.Query(
		ctx,
		blob.TimeRange{From: req.TimeRange.From, Through: req.TimeRange.To},
		blob.DefaultUser,
		req.Filter.Matchers...,
	)
	if err != nil {
		return nil, err
	}

	f, err := s.TempFile(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := storage.ExportDataset(ctx, s, f, blobs); err != nil {
		return nil, err
	}

	_, _ = f.Seek(0, 0)

	return httpx.Compose(
		httpx.WithContentType("application/tar+gzip"),
		httpx.WithMetadata(courier.Metadata{
			"Content-Disposition": {`attachment; filename="dataset.tar.gz"`},
		}),
	)(f), nil
}
