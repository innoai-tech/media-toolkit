package blob

import (
	"context"
	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	"github.com/innoai-tech/media-toolkit/pkg/blob"
	"github.com/innoai-tech/media-toolkit/pkg/filesize"
	"github.com/innoai-tech/media-toolkit/pkg/httputil"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"io"
	"net/http"
	"strconv"
)

func init() {
	BlobRouter.Register(courier.NewRouter(&GetBlob{}))
}

type GetBlob struct {
	httpx.MethodGet `path:"/blobs/:ref"`
	Ref             blob.RefString `name:"ref" in:"path"`
	Range           string         `name:"Range,omitempty" in:"header"`
}

func (req *GetBlob) Output(ctx context.Context) (any, error) {
	s := storage.StoreFromContext(ctx)
	info, err := s.Info(ctx, req.Ref.Ref())
	if err != nil {
		return nil, err
	}
	r, err := s.ReaderAt(ctx, info.Ref)
	if err != nil {
		return nil, err
	}

	// TODO should redirect when object enabled

	mediaType := "application/octet-stream"
	if info.Labels != nil {
		if mt, ok := info.Labels["_media_type"]; ok {
			mediaType = mt[0]
		}
	}

	if req.Range != "" {
		ranges, err := httputil.ParseRange(req.Range, r.Size())
		if err != nil {
			return nil, err
		}

		rng := ranges[0]

		if rng.Length > int64(50*filesize.KiB) {
			rng.Length = int64(50 * filesize.KiB)
		}

		return httpx.Compose(
			httpx.WithStatusCode(http.StatusPartialContent),
			httpx.WithContentType(mediaType),
			httpx.WithMetadata(courier.Metadata{
				"Cache-Control": {"max-age=31536000"},
				"Content-Range": {rng.ContentRange(r.Size())},
				"Accept-Ranges": {strconv.FormatInt(r.Size(), 10)},
			}),
		)(io.NewSectionReader(r, rng.Start, rng.Length)), nil
	}

	return httpx.Compose(
		httpx.WithContentType(mediaType),
		httpx.WithMetadata(courier.Metadata{
			"Cache-Control": {"max-age=31536000"},
			"Accept-Ranges": {strconv.FormatInt(r.Size(), 10)},
		}),
	)(io.NewSectionReader(r, 0, r.Size())), nil
}
