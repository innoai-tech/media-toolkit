package live

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/innoai-tech/media-toolkit/pkg/httputil"
	"github.com/innoai-tech/media-toolkit/webapp"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

type Player struct {
}

func (p *Player) Serve(ctx context.Context) error {
	l := logr.FromContextOrDiscard(ctx)

	router := mux.NewRouter()

	router.Path("/api").Methods(http.MethodGet).HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	router.PathPrefix("/").Handler(webapp.Handler)

	router.Use(httputil.LogHandler(l))

	s := &http.Server{}
	s.Addr = "0.0.0.0:777"
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

	return s.Shutdown(ctx)
}
