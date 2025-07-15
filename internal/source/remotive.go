package source

import (
	"fmt"
	"io"
	"time"

	"github.com/mmcdole/gofeed"
)

type RemotiveFeed struct {
	BaseSource
}

func NewRemotiveFeed(name, url string) RemotiveFeed {
	return RemotiveFeed{
		BaseSource{
			Source: name,
			Url:    url,
		},
	}
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
