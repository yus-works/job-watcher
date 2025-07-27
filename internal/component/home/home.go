package home

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/yus-works/job-watcher/internal/store"
)

func Register(tl *template.Template, st *store.JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
		defer cancel()

		jobs, err := st.GetJobs(ctx, req.URL.Query().Get("search"))
		if err != nil {
			http.Error(w, "timeout or db error", 500)
			log.Println(err)
			return
		}

		err = tl.ExecuteTemplate(w, "home", jobs)
		if err != nil {
			log.Println("ERROR: ", err)
		}
	}
}
