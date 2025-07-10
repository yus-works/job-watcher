package router

import (
	"html/template"
	"net/http"

	"github.com/yus-works/jod-watcher/internal/pages/home"
	"github.com/yus-works/jod-watcher/internal/store"
)

func RegisterHandlers(
	t *template.Template,
	s *store.JobStore,
) {
	http.HandleFunc("/", home.Register(t, s))
}
