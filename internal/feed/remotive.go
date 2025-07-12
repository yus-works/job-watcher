package feed

import (
	"fmt"
	"io"
	"time"

	"github.com/mmcdole/gofeed"
)

type RemotiveFeed struct {
	Url string
}

func (f RemotiveFeed) GetUrl() string {
	return f.Url
}

func (f RemotiveFeed) Parse(body io.Reader) ([]Item, error) {
	parser := gofeed.NewParser()

	feed, err := parser.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("ERROR: parsing feed")
	}

	out := make([]Item, 0, len(feed.Items))

	for _, fi := range feed.Items {
		when := time.Time{}
		if fi.PublishedParsed != nil {
			when = *fi.PublishedParsed
		} else if fi.UpdatedParsed != nil {
			when = *fi.UpdatedParsed
		}

		out = append(out, Item{
			Title:    fi.Title,
			Link:     fi.Link,
			Date:     when,

			Company:  fi.Custom["company"],
			Location: fi.Custom["location"],
			Kind:     fi.Custom["type"],
		})
	}

	return out, nil
}
