package fetcher

import (
	"context"
	"io"
	"log"
	"net/http"
	"sync"

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

func fetch(ctx context.Context, c *http.Client, feed feed.Feed) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feed.URL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

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
