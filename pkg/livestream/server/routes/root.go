package routes

import (
	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport"
	"github.com/innoai-tech/media-toolkit/pkg/livestream/server/routes/livestream"
)

var RootRouter = courier.NewRouter(httptransport.BasePath("/api"))

func init() {
	RootRouter.Register(livestream.LiveStreamRouter)
}
