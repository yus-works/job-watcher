package router

import (
	"context"
	"html/template"
	"net"
	"net/http"
	"time"

	"github.com/yus-works/job-watcher/internal/component/home"
	"github.com/yus-works/job-watcher/internal/component/jobs"
	"github.com/yus-works/job-watcher/internal/store"
)

func NewRouter(t *template.Template, s *store.JobStore) *http.ServeMux {
	mux := http.NewServeMux()

	fastResolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 500 * time.Millisecond}
			// ask Cloudflare
			return d.DialContext(ctx, network, "1.1.1.1:53")
		},
	}

	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			Resolver:  fastResolver,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:       100,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: false,
	}
	c := &http.Client{Transport: tr, Timeout: 10 * time.Second}

	registerFS(mux)
	registerHandlers(mux, t, s, c)

	return mux
}

func registerFS(m *http.ServeMux) {
	fs := http.FileServer(http.Dir("./static"))
	m.Handle("/static/",
		http.StripPrefix("/static/", fs),
	)
}

func registerHandlers(
	m *http.ServeMux,
	t *template.Template,
	s *store.JobStore,
	c *http.Client,
) {
	m.HandleFunc("/", home.Register(t, s))
	m.HandleFunc("/jobs", jobs.Register(t, c))
}
