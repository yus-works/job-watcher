package home

import (
	"html/template"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/yus-works/job-watcher/internal/store"
)

func Register(log *logrus.Entry, tl *template.Template, st *store.JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		err := tl.ExecuteTemplate(w, "home", nil)
		if err != nil {
			log.WithFields(logrus.Fields{
				"name": "home",
			}).WithError(err).Error("Failed to execute template")
		}
	}
}
