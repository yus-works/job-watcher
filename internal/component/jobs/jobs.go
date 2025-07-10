package jobs

import (
	"html/template"
	"net/http"

	"github.com/yus-works/job-watcher/internal/store"
)

func Register(tmpl *template.Template, st *store.JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		
	}
}
