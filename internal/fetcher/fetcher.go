package fetcher

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/yus-works/job-watcher/internal/feed"
)

func getItems(ctx context.Context, c *http.Client, feed feed.Feed) ([]feed.Item, error) {
	body, err := fetch(ctx, c, feed)
	if err != nil {
		log.Printf("Failed to fetch items (%s)", feed.Name)
	}

	defer body.Close()

	items, err := feed.Parse(feed, body)
	if err != nil {
		log.Printf("Failed to parse items (%s)", feed.Name)
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

	nt := &netTimings{startedAt: time.Now()}
	trace := &httptrace.ClientTrace{
		DNSStart:             func(httptrace.DNSStartInfo) { nt.startedAt = time.Now() },
		DNSDone:              func(httptrace.DNSDoneInfo) { nt.dns = time.Since(nt.startedAt) },
		ConnectStart:         func(_, _ string) { nt.startedAt = time.Now() },
		ConnectDone:          func(_, _ string, _ error) { nt.conn = time.Since(nt.startedAt) },
		TLSHandshakeStart:    func() { nt.startedAt = time.Now() },
		TLSHandshakeDone:     func(tls.ConnectionState, error) { nt.tls = time.Since(nt.startedAt) },
		GotFirstResponseByte: func() { nt.ttfb = time.Since(nt.startedAt); nt.ttfbAt = time.Now() },
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[%s] dns=%v conn=%v tls=%v ttfb=%v status=%d",
		f.Name, nt.dns, nt.conn, nt.tls, nt.ttfb, resp.StatusCode)

	return resp.Body, nil
}

func Stream(
	ctx context.Context,
	feeds []feed.Feed,
	client *http.Client,
) <-chan feed.Item {
	out := make(chan feed.Item, 64)

	var wg sync.WaitGroup

	// for each feed, spawn a getter routine
	for _, f := range feeds {
		feed := f // capture value

		wg.Add(1)

		go func() {
			defer wg.Done()

			items, err := getItems(ctx, client, feed)
			if err != nil {
				log.Printf("fetch %s: %v", feed.URL, err)
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
