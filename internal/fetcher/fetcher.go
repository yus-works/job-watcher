package fetcher

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/yus-works/job-watcher/internal/feed"
)

func fetch(ctx context.Context, c *http.Client, feed feed.Feed) ([]feed.Item, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feed.URL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return feed.Parse(feed, resp.Body)
}

func Stream(
	ctx context.Context,
	feeds []feed.Feed,
	client *http.Client,
) <-chan feed.Item {
	out := make(chan feed.Item, 64)

	var wg sync.WaitGroup

	// TODO: why is this loop ran twice? Is stream getting called twice?
	for _, f := range feeds {
		feed := f // capture value

		wg.Add(1)

		go func() {
			defer wg.Done()

			// TODO: move parsing to separate func call
			items, err := fetch(ctx, client, feed)
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
