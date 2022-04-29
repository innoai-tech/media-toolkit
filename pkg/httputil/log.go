package httputil

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

func LogHandler(l logr.Logger) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(rws http.ResponseWriter, req *http.Request) {
			nextCtx := req.Context()
			nextCtx = logr.NewContext(nextCtx, l)

			started := time.Now()

			rw := newLoggerResponseWriter(rws)

			handler.ServeHTTP(rw, req.WithContext(nextCtx))

			if !(req.URL.Path == "/") {
				values := []interface{}{
					"cost", time.Since(started),
					"status", rw.statusCode,
				}

				if ct := req.Header.Get("Content-Type"); ct != "" {
					values = append(values, "req.content-type", ct)
				}

				if rw.statusCode >= http.StatusInternalServerError {
					l.Error(errors.New("InternalError"), fmt.Sprintf("%s %s", req.Method, req.URL.String()), values...)
				} else {
					l.Info(fmt.Sprintf("%s %s", req.Method, req.URL.String()), values...)
				}

			}
		})
	}
}

func newLoggerResponseWriter(rw http.ResponseWriter) *loggerResponseWriter {
	h, hok := rw.(http.Hijacker)
	if !hok {
		h = nil
	}

	f, fok := rw.(http.Flusher)
	if !fok {
		f = nil
	}

	return &loggerResponseWriter{
		ResponseWriter: rw,
		Hijacker:       h,
		Flusher:        f,
	}
}

type loggerResponseWriter struct {
	http.ResponseWriter
	http.Hijacker
	http.Flusher

	headerWritten bool
	statusCode    int
	err           error
}

func (rw *loggerResponseWriter) WriteError(err error) {
	rw.err = err
}

func (rw *loggerResponseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *loggerResponseWriter) WriteHeader(statusCode int) {
	rw.writeHeader(statusCode)
}

func (rw *loggerResponseWriter) Write(data []byte) (int, error) {
	if rw.err == nil && rw.statusCode >= http.StatusBadRequest {
		rw.err = errors.New(string(data))
	}
	return rw.ResponseWriter.Write(data)
}

func (rw *loggerResponseWriter) writeHeader(statusCode int) {
	if !rw.headerWritten {
		rw.ResponseWriter.WriteHeader(statusCode)
		rw.statusCode = statusCode
		rw.headerWritten = true
	}
}
