package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/yus-works/job-watcher/cmd/middleware"
	"github.com/yus-works/job-watcher/internal/logging"
	"github.com/yus-works/job-watcher/internal/router"
	"github.com/yus-works/job-watcher/internal/store"
)

var LOG_LEVEL = os.Getenv("LOG_LEVEL")

func main() {

	logging.Init()

	switch LOG_LEVEL {
	case "DEBUG":
		logging.Level.Set(slog.LevelDebug)
	case "WARN":
		logging.Level.Set(slog.LevelWarn)
	case "ERROR":
		logging.Level.Set(slog.LevelError)
	default:
		logging.Level.Set(slog.LevelInfo)
	}

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

	fmt.Println("Listening on :8080")
	app := router.NewRouter(tmpl, store)
	handler := middleware.Chain(
		app,
		middleware.WithRequestLogger,
	)
	srv := &http.Server{Addr: ":8080", Handler: handler}
	log.Fatal(srv.ListenAndServe())
}
