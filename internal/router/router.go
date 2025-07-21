package router

import (
	"html/template"
	"net/http"

	"github.com/yus-works/job-watcher/internal/component/home"
	"github.com/yus-works/job-watcher/internal/component/jobs"
	"github.com/yus-works/job-watcher/internal/store"
)

func NewRouter(t *template.Template, s *store.JobStore) *http.ServeMux {
	mux := http.NewServeMux()

	registerFS(mux)
	registerHandlers(mux, t, s)

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
) {
	m.HandleFunc("/", home.Register(t, s))
	m.HandleFunc("/jobs", jobs.Register(t, s))
}
