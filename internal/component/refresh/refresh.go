package refresh

// TODO: move refresh out of comp its not a component
// TODO: make it so refresh can only be triggered once every 6 hours
import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yus-works/job-watcher/internal/component/jobs"
	"github.com/yus-works/job-watcher/internal/feed"
	"github.com/yus-works/job-watcher/internal/fetch"
	"github.com/yus-works/job-watcher/internal/perf"
	"github.com/yus-works/job-watcher/internal/registry"
	"github.com/yus-works/job-watcher/internal/store"
	"github.com/yus-works/job-watcher/internal/tmpl"
)

func sendTelegramNotification(
	log *logrus.Entry,
	ctx context.Context,
	tl *template.Template,
	jobItems []feed.JobItem,
	filterFunc func([]feed.JobItem) []feed.JobItem,
) error {
	botToken := os.Getenv("BOT_TOKEN")
	chatID := os.Getenv("CHAT_ID")

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	var message string
	var items []string
	for _, item := range filterFunc(jobItems) {
		itemStr, err := tmpl.Render(log, ctx, tl, "notification", jobs.NewDisplayItem(item))
		if err != nil {
			items = append(items, err.Error())
		}

		items = append(items, itemStr)
	}

	message = strings.Join(items, "\n\n")

	payload := map[string]any{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "HTML",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func Register(
	log *logrus.Entry,
	tl *template.Template,
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

		ctx := req.Context()

		if last, err := st.LastRefresh(ctx); err == nil && !last.IsZero() {
			if wait := 3*time.Hour - time.Since(last); wait > 0 {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprintf(w, "too frequent; try again in %s\n", wait)
				return
			}
		}

		itemsCh := fetch.Stream(log, ctx, registry.FEEDS, cl)

		stopTotal := perf.StartTimer(log, logrus.DebugLevel, "jobs_total")
		defer stopTotal(nil)

		insertedCount := 0
		skippedCount := 0

		items := make([]feed.JobItem, len(itemsCh))

		// TODO: have this writing happen on a dedicated goroutine myb
		for it := range itemsCh {
			inserted, err := st.Insert(log, ctx, it)
			if err != nil {
				msg := fmt.Sprintf("insert failed: %v", err)

				log.WithError(err).Warn("insert item")

				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, msg)
				return
			}

			if inserted {
				insertedCount++

				items = append(items, it)
			} else {
				skippedCount++
			}
		}

		_ = st.SetLastRefresh(ctx, time.Now())

		filterFn := func(items []feed.JobItem) []feed.JobItem {
			var finalItems []feed.JobItem
			for _, it := range items {
				if strings.Contains(strings.ToLower(it.Title), "senior") {
					continue
				}

				if slices.Contains(it.Tags, "senior") {
					continue
				}

				goRegex := regexp.MustCompile(`(?i)\bgo\b|\bgolang\b`)
				if !goRegex.MatchString(it.Title) && !goRegex.MatchString(it.Company) {
					continue
				}

				finalItems = append(finalItems, it)
			}

			return finalItems
		}

		sendTelegramNotification(log, ctx, tl, items, filterFn)

		st.DeleteOldPostings(ctx)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(
			w,
			"fetched %d jobs, inserted %d, skipped %d\n",
			insertedCount+skippedCount, insertedCount, skippedCount,
		)
	}
}
