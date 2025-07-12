package fetcher

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

func fetch(ctx context.Context, c *http.Client, url string) ([]Item, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return decodeItems(resp.Body)
}

func Stream(
	ctx context.Context,
	urls []string,
) <-chan Item {
	out := make(chan Item)

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       100,
			IdleConnTimeout:    90 * time.Second,
			DisableCompression: false,
		},
	}

	var wg sync.WaitGroup

	for _, u := range urls {
		url := u // capture value

		wg.Add(1)

		go func() {
			defer wg.Done()

			items, err := fetch(ctx, client, url)
			if err != nil {
				log.Printf("fetch %s: %v", url, err)
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
