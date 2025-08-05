// cmd/job-watcher/main.go
package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/yus-works/job-watcher/internal/logging"
	"github.com/yus-works/job-watcher/internal/router"
	"github.com/yus-works/job-watcher/internal/store"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/job-watcher.db" // local
	}

	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = "data/job-watcher.log" // local
	}

	// ensure parents exist
	_ = os.MkdirAll(filepath.Dir(dbPath), 0o755)
	_ = os.MkdirAll(filepath.Dir(logPath), 0o755)

	logg, closer, err := logging.New(logrus.InfoLevel, logPath)
	if err != nil {
		log.Fatalf("logger init: %v", err)
	}
	defer closer.Close()

	st, err := store.NewJobStore(dbPath)
	if err != nil {
		logg.WithError(err).Fatal("open db")
	}
	defer st.Close()

	err = st.CreateTables(context.Background())
	if err != nil {
		logg.WithError(err).Fatal("create/check tables")
		return
	}

	tl := template.Must(template.ParseGlob("internal/tmpl/*.html"))
	mux := router.NewRouter(tl, st, logg)

	logg.WithField("addr", ":8080").Info("listening")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		logg.WithError(err).Fatal("server stopped")
	}
}
