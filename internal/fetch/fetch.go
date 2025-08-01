package fetch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/yus-works/job-watcher/internal/feed"
	"github.com/yus-works/job-watcher/internal/logging"
	"github.com/yus-works/job-watcher/internal/perf"
)

func getItems(ctx context.Context, c *http.Client, feed feed.Feed) ([]feed.JobItem, error) {
	body, err := fetch(ctx, c, feed)
	if err != nil {
		logging.From(ctx).Error("Failed to fetch items", "source", feed.Name)
	}

	defer body.Close()

	items, err := feed.Parse(ctx, feed, body)
	if err != nil {
		logging.From(ctx).Error("Failed to parse items", "source", feed.Name)
	}

	return items, nil
}

type netTimings struct {
	startedAt, ttfbAt    time.Time
	dns, conn, tls, ttfb time.Duration
}

func fetch(ctx context.Context, c *http.Client, f feed.Feed) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "JobWatcher/0.1 (+https://example.com)")

	// attach httptrace only if debug enabled
	trace, done := perf.NetTrace(
		ctx,
		slog.LevelDebug,
		"fetch",
		slog.String("feed", f.Name),
		slog.String("url", f.URL),
	)
	if trace != nil {
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		resp.Body.Close()

		done(resp.StatusCode) // no-op if perf logging is off
		return nil, fmt.Errorf("%s: status %d: %q", f.Name, resp.StatusCode, bytes.TrimSpace(snippet))
	}

	done(resp.StatusCode)
	return resp.Body, nil
}

func Stream(
	ctx context.Context,
	feeds []feed.Feed,
	client *http.Client,
) <-chan feed.JobItem {
	out := make(chan feed.JobItem, 64)

	var wg sync.WaitGroup

	// for each feed, spawn a getter routine
	for _, f := range feeds {
		feed := f // capture value

		wg.Add(1)

		go func() {
			defer wg.Done()

			items, err := getItems(ctx, client, feed)
			if err != nil {
				logging.From(ctx).Warn("Failed to fetch", "url", feed.URL, "err", err)
				return
			}

			for _, it := range items {
				select {
				case out <- it: // runs if out is ready to receive
				case <-ctx.Done(): // runs if ctx.Done is ready to send
					return
				}
			}
		}()
	}

	// closer
	go func() {
		wg.Wait()
		defer close(out)
	}()

	return out
}
