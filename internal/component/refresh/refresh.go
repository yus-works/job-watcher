package refresh

// TODO: move refresh out of comp its not a component
// TODO: make it so refresh can only be triggered once every 6 hours
import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yus-works/job-watcher/internal/fetch"
	"github.com/yus-works/job-watcher/internal/perf"
	"github.com/yus-works/job-watcher/internal/registry"
	"github.com/yus-works/job-watcher/internal/store"
)

var lastRun time.Time

func Register(
	log *logrus.Entry,
	st *store.JobStore,
	cl *http.Client,
	secret string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if secret == "" {
			http.NotFound(w, req)
			return
		}

		got := req.Header.Get("Authorization")
		if after, ok := strings.CutPrefix(got, "Bearer "); ok {
			got = after
		} else {
			got = req.Header.Get("X-Refresh-Secret")
		}

		// constant-time compare to avoid timing leaks ?
		if subtle.ConstantTimeCompare([]byte(got), []byte(secret)) != 1 {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(w, "forbidden")
			return
		}

		// naive cool-down: block if run < 4h ago
		if !lastRun.IsZero() && time.Since(lastRun) < 4*time.Hour {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintf(w, "too frequent; try again in %s\n", (6*time.Hour)-time.Since(lastRun))
			return
		}
		lastRun = time.Now()

		w.Header().Set("Content-Type", "text/plain")
		ctx := req.Context()

		itemsCh := fetch.Stream(log, ctx, registry.FEEDS, cl)

		stopTotal, _ := perf.StartTimer(log, logrus.DebugLevel, "jobs_total")
		defer stopTotal(nil) // TODO: this might break idk

		insertedCount := 0
		skippedCount := 0

		for it := range itemsCh {
			inserted, err := st.Insert(log, ctx, it)
			if err != nil {
				msg := fmt.Sprintf("insert failed: %v", err)

				log.WithFields(logrus.Fields{
					"err": err,
				}).Warn("insert item")

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
