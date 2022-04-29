package live

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/innoai-tech/media-toolkit/pkg/httputil"
	"github.com/innoai-tech/media-toolkit/pkg/livestream"
	"github.com/innoai-tech/media-toolkit/pkg/livestream/server"
)

type StreamPlayer struct {
	Addr    string
	Streams []livestream.Stream
}

func (p *StreamPlayer) Serve(ctx context.Context) error {
	l := logr.FromContextOrDiscard(ctx)

	router := mux.NewRouter()

	lvs := server.NewLiveStreamServer(p.Streams)

	router.PathPrefix("/api").Handler(lvs.Handler())
	router.PathPrefix("/").Handler(WebUI)

	router.Use(httputil.LogHandler(l))

	s := &http.Server{}
	s.Addr = p.Addr
	s.Handler = router

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		l.Info(fmt.Sprintf(
			"live player serve on http://%s (%s/%s)",
			s.Addr,
			runtime.GOOS, runtime.GOARCH,
		))

		if e := s.ListenAndServe(); e != nil {
			l.Error(e, "")
		}
	}()

	<-stopCh

	timeout := 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		_ = lvs.Shutdown(ctx)
	}()

	return s.Shutdown(ctx)
}
