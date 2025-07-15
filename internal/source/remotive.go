package source

import (
	"fmt"
	"io"
	"time"

	"github.com/mmcdole/gofeed"
)

type RemotiveFeed struct {
	Source string
	Url    string
}

func (f RemotiveFeed) GetUrl() string {
	return f.Url
}

func (f RemotiveFeed) GetSource() string {
	return f.Source
}

func (f RemotiveFeed) Parse(body io.Reader) ([]Item, error) {
	parser := gofeed.Parser()

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
			Source: f.GetSource(),
			Title:  fi.Title,
			Link:   fi.Link,
			Date:   when,

			Company:  fi.Custom["company"],
			Location: fi.Custom["location"],
			Kind:     fi.Custom["type"],
		})
	}

	return out, nil
}
