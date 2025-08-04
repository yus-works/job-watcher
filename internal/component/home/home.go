package home

import (
	"html/template"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yus-works/job-watcher/internal/store"
)

func Register(log *logrus.Entry, tl *template.Template, st *store.JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		last, _ := st.LastRefresh(req.Context())
		data := struct {
			LastRefreshISO string
			SSEURL         string
		}{
			LastRefreshISO: func() string {
				if last.IsZero() {
					return ""
				}
				return last.UTC().Format(time.RFC3339Nano)
			}(),
			SSEURL: "/jobs/stream",
		}

		if err := tl.ExecuteTemplate(w, "home", data); err != nil {
			log.WithFields(logrus.Fields{
				"name": "home",
			}).WithError(err).Error("Failed to execute template")
		}
	}
}
