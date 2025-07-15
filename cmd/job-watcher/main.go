package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/yus-works/job-watcher/internal/router"
	"github.com/yus-works/job-watcher/internal/store"
)

func clockSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// make sure writer supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			msg := fmt.Sprintf("event: clockEvent\ndata: %s\n\n", t.Format(time.RFC3339))

			fmt.Fprint(w, msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

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

	http.HandleFunc("/clock", clockSSE)

	fmt.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
