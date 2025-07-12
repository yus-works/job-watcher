package fetcher

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/yus-works/job-watcher/internal/feed"
)

func fetch[T feed.Feed](ctx context.Context, c *http.Client, feed T) ([]feed.Item, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feed.GetUrl(), nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return feed.Parse(resp.Body)
}

func Stream[T feed.Feed](
	ctx context.Context,
	feeds []T,
) <-chan feed.Item {
	out := make(chan feed.Item)

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       100,
			IdleConnTimeout:    90 * time.Second,
			DisableCompression: false,
		},
	}

	var wg sync.WaitGroup

	for _, f := range feeds {
		feed := f // capture value

		wg.Add(1)

		go func() {
			defer wg.Done()

			items, err := fetch(ctx, client, feed)
			if err != nil {
				log.Printf("fetch %s: %v", feed.GetUrl(), err)
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
