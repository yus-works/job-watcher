package home

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/yus-works/job-watcher/internal/store"
)

func seed(ctx context.Context, st *store.JobStore) {
	for i := 1; i <= 10; i++ {
		j := store.Job{
			ID:      fmt.Sprintf("test-%02d", i),
			Title:   fmt.Sprintf("Dummy Job #%d", i),
			URL:     fmt.Sprintf("https://example.com/test/%d", i),
			Company: fmt.Sprintf("ExampleCorp %d", i),
		}
		if err := st.Insert(ctx, j); err != nil {
			log.Printf("insert %v: %v", j, err)
		}
	}

	fmt.Println("SEEDING DONE")
}

func Register(tmpl *template.Template, st *store.JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// TODO: figure out why this block is ran twice

		ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
		defer cancel()

		err := st.CreateTables(ctx)
		if err != nil {
			http.Error(w, "failed to create jobs table", 500)
			log.Println(err)
			return
		}

		jobs, err := st.GetJobs(ctx, req.URL.Query().Get("search"))
		if err != nil {
			http.Error(w, "timeout or db error", 500)
			log.Println(err)
			return
		}

		err = tmpl.ExecuteTemplate(w, "home", jobs)
		if err != nil {
			log.Println("ERROR: ", err)
		}
	}
}
