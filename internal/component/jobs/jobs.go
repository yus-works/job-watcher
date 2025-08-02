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

func Register(
	log *logrus.Entry,
	tl *template.Template,
	st *store.JobStore,
	cl *http.Client,
) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// feeds := []string{
		// 	"https://weworkremotely.com/categories/remote-programming-jobs.rss",
		// 	"http://rss.infostud.com/poslovi/",
		// 	"https://profession.hu/allasok?rss",
		// 	"https://mernokallasok.hu/rss_friss.xml",
		// }

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
		stopTotal, _ := perf.StartTimer(log, logrus.DebugLevel, "jobs_total")
		defer stopTotal(map[string]any{
			"test1": "idk",
			"test2": "bruh",
		})

		stopTTFI, _ := perf.StartTimer(log, logrus.DebugLevel, "ttfi")
		first := true

		jobsCh, errsCh := st.StreamJobs(log, ctx, "")

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

				stopRender, _ := perf.StartTimer(log, logrus.DebugLevel, "render")
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

				log.WithFields(logrus.Fields{
					"err": err,
				}).Error("fetch jobs")
			}
		}
	}
}
