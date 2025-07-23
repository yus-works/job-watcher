package jobs

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/yus-works/job-watcher/internal/fetcher"
	"github.com/yus-works/job-watcher/internal/registry"
	"github.com/yus-works/job-watcher/internal/store"
	"github.com/yus-works/job-watcher/internal/tmpl"
)

func Register(tl *template.Template, st *store.JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// apis := []string{
		// 	"https://www.arbeitnow.com/api/job-board-api",
		// }
		//
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

		client := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       100,
				IdleConnTimeout:    90 * time.Second,
				DisableCompression: false,
			},
		}

		itemsCh := fetcher.Stream(ctx, registry.FEEDS, client)

		for {
			select {
			case it, ok := <-itemsCh:
				if !ok {
					// all jobs sent, tell the client we're done
					fmt.Fprintf(w, "event: done\ndata: bye\n\n")
					flusher.Flush()
					return
				}

				card, err := tmpl.Render(tl, "card", NewDisplayItem(it))
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
