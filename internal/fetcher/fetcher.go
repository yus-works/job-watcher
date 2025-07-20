package fetcher

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/yus-works/job-watcher/internal/feed"
)

func parse(currFeed feed.Feed, body io.Reader) ([]feed.Item, error) {
	parser := gofeed.NewParser()

	items, err := parser.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("ERROR: parsing feed")
	}

	out := make([]feed.Item, 0, len(items.Items))

	for _, fi := range items.Items {
		when := time.Time{}
		if fi.PublishedParsed != nil {
			when = *fi.PublishedParsed
		} else if fi.UpdatedParsed != nil {
			when = *fi.UpdatedParsed
		}

		title := currFeed.Mapping.Title
		if fi.Title != "" {
			title = fi.Title
		}

		link := currFeed.Mapping.Link
		if fi.Link != "" {
			link = fi.Link
		}

		out = append(out, feed.Item{
			Source:   currFeed.Name,
			Title:    title,
			Link:     link,
			Company:  fi.Custom[currFeed.Mapping.Company],
			Location: fi.Custom[currFeed.Mapping.Location],
			Kind:     fi.Custom[currFeed.Mapping.Kind],
			Date:     when,
		})
	}

	return out, nil
}

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

	return parse(feed, resp.Body)
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
