package refresh

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/yus-works/job-watcher/internal/fetch"
	"github.com/yus-works/job-watcher/internal/logging"
	"github.com/yus-works/job-watcher/internal/perf"
	"github.com/yus-works/job-watcher/internal/registry"
	"github.com/yus-works/job-watcher/internal/store"
)

func Register(
	s *store.JobStore,
	c *http.Client,
) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		ctx := req.Context()

		itemsCh := fetch.Stream(ctx, registry.FEEDS, c)

		stopTotal, _ := perf.StartTimer(ctx, slog.LevelDebug, "jobs_total")
		defer stopTotal()

		insertedCount := 0
		skippedCount := 0

		for it := range itemsCh {
			inserted, err := s.Insert(ctx, store.FromFeedItem(it))
			if err != nil {
				msg := fmt.Sprintf("insert failed: %v", err)

				logging.From(ctx).Warn(msg)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, msg)
				return
			}

			if inserted {
				insertedCount++
			} else {
				skippedCount++
			}
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(
			w,
			"fetched %d jobs, inserted %d, skipped %d\n",
			insertedCount+skippedCount, insertedCount, skippedCount,
		)
	}
}
