package server

import (
	"context"
	"github.com/go-courier/httptransport"
	"github.com/innoai-tech/media-toolkit/pkg/livestream/server/routes"
	"github.com/innoai-tech/media-toolkit/pkg/storage"
	"github.com/innoai-tech/media-toolkit/pkg/storage/config"
	"github.com/innoai-tech/media-toolkit/pkg/version"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"net/http"
	"sort"

	"github.com/innoai-tech/media-toolkit/pkg/livestream"
)

func NewLiveStreamServer(streams []livestream.Stream) *LiveStreamServer {
	hub := livestream.NewStreamHub()

	for i := range streams {
		hub.AddStream(streams[i])
	}

	// TODO
	store, _ := storage.New(config.DefaultConfig)

	return &LiveStreamServer{
		hub:   hub,
		store: store,
	}
}

type LiveStreamServer struct {
	hub   *livestream.StreamHub
	store storage.Store
}

func (ls *LiveStreamServer) Shutdown(ctx context.Context) error {
	ls.store.Stop()
	return nil
}

func (ls *LiveStreamServer) Handler() http.Handler {
	return httptransport.MiddlewareChain(
		cors.Default().Handler,
		ls.injectContext,
	)(ls.apis())
}

func (ls *LiveStreamServer) injectContext(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		ctx = livestream.NewContextWithStreamHub(ctx, ls.hub)
		ctx = storage.NewContextWithStore(ctx, ls.store)

		handler.ServeHTTP(rw, req.WithContext(ctx))
	})
}

func (ls *LiveStreamServer) apis() http.Handler {
	allRoutes := routes.RootRouter.Routes()

	routeMetas := make([]*httptransport.HttpRouteMeta, len(allRoutes))
	for i := range allRoutes {
		routeMetas[i] = httptransport.NewHttpRouteMeta(allRoutes[i])
	}

	httpRouter := httprouter.New()

	sort.Slice(routeMetas, func(i, j int) bool {
		return routeMetas[i].Key() < routeMetas[j].Key()
	})

	for i := range routeMetas {
		httpRoute := routeMetas[i]
		httpRoute.Log()

		httpRouter.HandlerFunc(
			httpRoute.Method(),
			httpRoute.Path(),
			httptransport.NewHttpRouteHandler(
				&httptransport.ServiceMeta{
					Name:    "livestream",
					Version: version.FullVersion(),
				},
				httpRoute,
				httptransport.NewRequestTransformerMgr(nil, nil),
			).ServeHTTP,
		)
	}

	return httpRouter
}
