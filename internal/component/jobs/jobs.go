package jobs

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/yus-works/job-watcher/internal/fetcher"
	"github.com/yus-works/job-watcher/internal/source"
	"github.com/yus-works/job-watcher/internal/store"
	"github.com/yus-works/job-watcher/internal/tmpl"
)

func Register(tl *template.Template, st *store.JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// apis := []string{
		// 	"https://remotive.com/api/remote-jobs?category=software-dev",
		// 	"https://remoteok.com/api",
		// 	"https://jobicy.com/api/v2/remote-jobs?count=100&geo=europe&industry=engineering&tag=Golang",
		// 	"https://himalayas.app/jobs/api",
		// 	"https://www.arbeitnow.com/api/job-board-api",
		// }
		//
		// feeds := []string{
		// 	"https://remotive.com/remote-jobs/feed/software-dev",
		// 	"https://remoteok.com/remote-jobs.rss",
		// 	"https://jobicy.com/feed/job_feed",
		// 	"https://himalayas.app/jobs/rss",
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

		feeds := make([]source.RemotiveFeed, 0)
		feeds = append(feeds,
			source.NewRemotiveFeed(
				"Remotive", "http://localhost:8000/remotive.rss",
			),
		)

		client := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       100,
				IdleConnTimeout:    90 * time.Second,
				DisableCompression: false,
			},
		}

		itemsCh := fetcher.Stream(ctx, feeds, client)

		for {
			select {
			case it, ok := <-itemsCh:
				if !ok {
					// all jobs sent, tell the client we're done
					fmt.Fprintf(w, "event: done\ndata: bye\n\n")
					flusher.Flush()
					return
				}

				html, err := tmpl.Render(tl, "card", it)
				if err != nil {
					fmt.Fprintf(w, "event: renderFailed\ndata: %s\n\n", html)
					flusher.Flush()

					fmt.Fprint(w, "event: done\ndata: bye\n\n")
					flusher.Flush()
					return
				}

				fmt.Fprintf(
					w,
					"event: foundJobs\ndata: %s\n\n",
					strings.ReplaceAll(html, "\n", ""),
				)
				flusher.Flush()

			// client hung-up or timed-out
			case <-ctx.Done():
				return
			}
		}
	}
}
