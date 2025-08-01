package home

import (
	"context"
	"html/template"
	"net/http"
	"time"

	"github.com/yus-works/job-watcher/internal/logging"
	"github.com/yus-works/job-watcher/internal/store"
)

func Register(tl *template.Template, st *store.JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
		defer cancel()

		jobs, err := st.GetJobs(ctx, req.URL.Query().Get("search"))
		if err != nil {
			http.Error(w, "timeout or db error", 500)
			logging.From(ctx).Error("Failed to get jobs", "err", err)
			return
		}

		err = tl.ExecuteTemplate(w, "home", jobs)
		if err != nil {
			logging.From(ctx).Error(
				"Failed to execute template",
				"name", "home",
				"err", err,
			)
		}
	}
}
