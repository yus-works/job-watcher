package jobs

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"

	"github.com/yus-works/job-watcher/internal/fetch"
	"github.com/yus-works/job-watcher/internal/perf"
	"github.com/yus-works/job-watcher/internal/registry"
	"github.com/yus-works/job-watcher/internal/store"
	"github.com/yus-works/job-watcher/internal/tmpl"
)

func Register(
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

		itemsCh := fetch.Stream(ctx, registry.FEEDS, cl)

		// timers — no time.Now(); they no-op when debug/info not enabled
		stopTotal, _ := perf.StartTimer(ctx, slog.LevelDebug, "jobs_total")
		defer stopTotal()

		stopTTFI, _ := perf.StartTimer(ctx, slog.LevelDebug, "ttfi")
		first := true

		for {
			select {
			case it, ok := <-itemsCh:
				if !ok {
					// all jobs sent, tell the client we're done
					fmt.Fprintf(w, "event: done\ndata: bye\n\n")
					flusher.Flush()
					return
				}

				if first {
					first = false
					stopTTFI()
				}

				stopRender, _ := perf.StartTimer(ctx, slog.LevelDebug, "render")
				card, err := tmpl.Render(tl, "card", NewDisplayItem(it))
				stopRender()

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

			// client hung-up or timed-out
			case <-ctx.Done():
				return
			}
		}
	}
}
