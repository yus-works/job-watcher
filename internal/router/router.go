package router

import (
	"context"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yus-works/job-watcher/internal/component/home"
	"github.com/yus-works/job-watcher/internal/component/jobs"
	"github.com/yus-works/job-watcher/internal/component/refresh"
	"github.com/yus-works/job-watcher/internal/store"
)

func NewRouter(
	t *template.Template,
	s *store.JobStore,
	l *logrus.Logger,
) *http.ServeMux {
	m := http.NewServeMux()

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

	registerFS(m)
	registerHandlers(l, m, t, s, c)

	return m
}

func registerFS(m *http.ServeMux) {
	fs := http.FileServer(http.Dir("./static"))
	m.Handle("/static/",
		http.StripPrefix("/static/", fs),
	)
}

func registerHandlers(
	l logrus.FieldLogger,
	m *http.ServeMux,
	t *template.Template,
	s *store.JobStore,
	c *http.Client,
) {
	mkLogger := func(r *http.Request) *logrus.Entry {
		return l.WithFields(logrus.Fields{
			"path": r.URL.Path,
		})
	}

	refreshSecret := os.Getenv("REFRESH_TOKEN")

	// TODO: too much voodoo
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		home.Register(mkLogger(r), t, s)(w, r)
	})
	m.HandleFunc("/jobs/stream", func(w http.ResponseWriter, r *http.Request) {
		jobs.Register(mkLogger(r), t, s, c)(w, r)
	})
	m.HandleFunc("/jobs/reset", func(w http.ResponseWriter, r *http.Request) {
		jobs.Register(mkLogger(r), t, s, c)(w, r)
	})
	m.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		refresh.Register(mkLogger(r), s, c, refreshSecret)(w, r)
	})
}
