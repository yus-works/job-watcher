package home

import (
	"context"
	"html/template"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yus-works/job-watcher/internal/store"
)

func Register(log *logrus.Entry, tl *template.Template, st *store.JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
		defer cancel()

		// TODO: this seems redundant
		jobs, err := st.GetJobs(log, ctx, req.URL.Query().Get("search"))
		if err != nil {
			http.Error(w, "timeout or db error", 500)
			log.WithFields(logrus.Fields{
				"err": err,
			}).Error("get jobs")
			return
		}

		err = tl.ExecuteTemplate(w, "home", jobs)
		if err != nil {
			log.WithFields(logrus.Fields{
				"name": "home",
				"err":  err,
			}).Error("Failed to execute template")
		}
	}
}
