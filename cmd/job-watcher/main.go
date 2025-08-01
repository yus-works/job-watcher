// cmd/job-watcher/main.go
package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/yus-works/job-watcher/internal/logging"
	"github.com/yus-works/job-watcher/internal/router"
	"github.com/yus-works/job-watcher/internal/store"
)

func main() {
	logg, closer, err := logging.New(logrus.DebugLevel, "job-watcher.log")

	if err != nil {
		log.Fatalf("logger init: %v", err)
	}
	defer closer.Close()

	tl := template.Must(template.ParseGlob("internal/tmpl/*.html"))
	st, err := store.NewJobStore("job-store.db")
	if err != nil {
		logg.WithError(err).Fatal("open db")
	}
	defer st.Close()

	mux := router.NewRouter(tl, st, logg)

	logg.WithField("addr", ":8080").Info("listening")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		logg.WithError(err).Fatal("server stopped")
	}
}
