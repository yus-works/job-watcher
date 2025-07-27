package middleware

import (
	"log/slog"
	"net/http"

	"github.com/rs/xid"
	"github.com/yus-works/job-watcher/internal/logging"
)

type Middleware = func(http.Handler) http.Handler

func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func WithRequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := xid.New().String()
		l := slog.Default().With("req_id", reqID, "path", r.URL.Path)
		ctx := logging.Into(r.Context(), l)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
