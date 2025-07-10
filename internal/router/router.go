package router

import (
	"html/template"
	"net/http"

	"github.com/yus-works/job-watcher/internal/component/home"
	"github.com/yus-works/job-watcher/internal/component/jobs"
	"github.com/yus-works/job-watcher/internal/store"
)

func RegisterHandlers(
	t *template.Template,
	s *store.JobStore,
) {
	http.HandleFunc("/", home.Register(t, s))
	http.HandleFunc("/jobs", jobs.Register(t, s))
}
