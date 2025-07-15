package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/yus-works/job-watcher/internal/router"
	"github.com/yus-works/job-watcher/internal/store"
)

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/",
		http.StripPrefix("/static/", fs),
	)

	store, err := store.NewJobStore("job-store.db")
	if err != nil {
		log.Fatal("Failed to open db")
	}

	err = store.CreateTables(context.Background())
	if err != nil {
		log.Println(err)
		return
	}

	tmpl := template.Must(template.ParseGlob("internal/tmpl/*.html"))
	router.RegisterHandlers(tmpl, store)

	fmt.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
