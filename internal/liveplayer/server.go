package liveplayer

import (
	"context"
	"fmt"
	"github.com/innoai-tech/media-toolkit/pkg/livestream/core"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/innoai-tech/media-toolkit/pkg/httputil"
	"github.com/innoai-tech/media-toolkit/pkg/livestream/server"
)

type StreamPlayer struct {
	Addr    string
	Streams []core.Stream
}

func (p *StreamPlayer) Serve(ctx context.Context) error {
	l := logr.FromContextOrDiscard(ctx)

	router := mux.NewRouter()

	lvs := server.NewLiveStreamServer(ctx, p.Streams)

	router.PathPrefix("/api").Handler(lvs.Handler())
	router.PathPrefix("/").Handler(WebUI)

	router.Use(gorillaHandlers.CompressHandler)
	router.Use(httputil.LogHandler(l))

	s := &http.Server{}
	s.Addr = p.Addr
	s.Handler = router

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		l.Info(
			fmt.Sprintf("live player serve on http://%s (%s/%s)",
				s.Addr,
				runtime.GOOS, runtime.GOARCH),
		)

		if e := s.ListenAndServe(); e != nil {
			l.Error(e, "")
		}
	}()

	<-stopCh

	timeout := 10 * time.Second

	ctx, cancel := context.WithTimeout(logr.NewContext(context.Background(), l), timeout)
	defer cancel()

	wg := sync.WaitGroup{}

	for _, canShutdown := range []CanShutdown{lvs, s} {
		wg.Add(1)
		go func(canShutdown CanShutdown) {
			defer wg.Done()
			if err := canShutdown.Shutdown(ctx); err != nil {
				l.Error(err, "Shutdown")
			}
		}(canShutdown)
	}

	wg.Wait()
	return nil
}

type CanShutdown interface {
	Shutdown(ctx context.Context) error
}
