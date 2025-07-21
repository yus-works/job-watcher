package fetcher

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/yus-works/job-watcher/internal/feed"
)

func HumanAgeGreedy(dur time.Duration) string {
	if dur <= 0 {
		return "0h"
	}

	const (
		hour       = time.Hour
		dayHours   = 24 * hour
		weekHours  = 7 * dayHours
		monthHours = 30 * dayHours
	)

	months := dur / monthHours
	dur -= months * monthHours

	weeks := dur / weekHours
	dur -= weeks * weekHours

	days := dur / dayHours
	dur -= days * dayHours

	hours := dur / hour

	parts := make([]string, 0, 4)
	if months > 0 {
		parts = append(parts, fmt.Sprintf("%dmo", months))
	}
	if weeks > 0 {
		parts = append(parts, fmt.Sprintf("%dw", weeks))
	}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}

	if len(parts) == 0 {
		return "0h"
	}
	return strings.Join(parts, " ")
}

func fetch(ctx context.Context, c *http.Client, feed feed.Feed) ([]feed.Item, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feed.URL(), nil)
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

func Stream(
	ctx context.Context,
	feeds []feed.Feed,
	client *http.Client,
) <-chan feed.Item {
	out := make(chan feed.Item, 64)

	var wg sync.WaitGroup

	for _, f := range feeds {
		feed := f // capture value

		wg.Add(1)

		go func() {
			defer wg.Done()

			// TODO: move parsing to separate func call
			items, err := fetch(ctx, client, feed)
			if err != nil {
				log.Printf("fetch %s: %v", feed.URL(), err)
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
