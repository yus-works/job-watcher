package refresh

// TODO: move refresh out of comp its not a component
import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/yus-works/job-watcher/internal/fetch"
	"github.com/yus-works/job-watcher/internal/perf"
	"github.com/yus-works/job-watcher/internal/registry"
	"github.com/yus-works/job-watcher/internal/store"
)

func Register(
	log *logrus.Entry,
	st *store.JobStore,
	cl *http.Client,
) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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
