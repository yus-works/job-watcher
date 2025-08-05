package jobs

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yus-works/job-watcher/internal/perf"
	"github.com/yus-works/job-watcher/internal/store"
	"github.com/yus-works/job-watcher/internal/tmpl"
)

func StreamJobs(
	log *logrus.Entry,
	tl *template.Template,
	st *store.JobStore,
	cl *http.Client,
) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		ctx := req.Context()

		// timers — no time.Now(); they no-op when debug/info not enabled
		stopTotal := perf.StartTimer(log, logrus.DebugLevel, "jobs_total")
		defer stopTotal(map[string]any{
			"test1": "idk",
			"test2": "bruh",
		})

		stopTTFI := perf.StartTimer(log, logrus.DebugLevel, "ttfi")
		first := true

		filter := req.URL.Query().Get("search")
		sort := req.URL.Query().Get("sort")

		jobsCh, errsCh := st.StreamJobs(log, ctx, filter, sort)

		for jobsCh != nil || errsCh != nil {
			select {
			case job, ok := <-jobsCh:
				if !ok {
					jobsCh = nil

					// all jobs sent, tell the client we're done
					fmt.Fprintf(w, "event: done\ndata: bye\n\n")
					flusher.Flush()
					return
				}

				if first {
					first = false
					stopTTFI(map[string]any{
						"test1": "idk",
						"test2": "bruh",
					})
				}

				stopRender := perf.StartTimer(log, logrus.DebugLevel, "render")
				card, err := tmpl.Render(log, ctx, tl, "card", NewDisplayItem(job))
				stopRender(nil)

				if err != nil {
					fmt.Fprintf(w, "event: renderFailed\ndata: %s\n\n", card)
					flusher.Flush()

					fmt.Fprint(w, "event: done\ndata: bye\n\n")
					flusher.Flush()
					return
				}

				fmt.Fprintf(
					w,
					"event: foundJobs\ndata: %s\n\n",
					strings.ReplaceAll(card, "\n", ""),
				)
				flusher.Flush()
			case err, ok := <-errsCh:
				if !ok {
					errsCh = nil
					continue
				}

				log.WithError(err).Error("fetch jobs")
			}
		}
	}
}

func ResetJobs(
	log *logrus.Entry,
	tl *template.Template,
	st *store.JobStore,
	cl *http.Client,
) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		data := struct {
			SSEURL string
		}{
			SSEURL: "/jobs/stream?" +
				"search=" + req.URL.Query().Get("search") +
				"&" +
				"sort=" + req.URL.Query().Get("sort"),
		}

		if err := tl.ExecuteTemplate(w, "jobs", data); err != nil {
			log.WithFields(logrus.Fields{
				"name": "jobs",
			}).WithError(err).Error("Failed to execute template")
		}
	}
}
