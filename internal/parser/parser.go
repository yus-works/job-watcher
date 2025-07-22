package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/yus-works/job-watcher/internal/feed"
)

func ParseJSON(curr feed.Feed, objs []map[string]json.RawMessage) ([]feed.Item, error) {
	out := make([]feed.Item, 0, len(objs))
	now := time.Now()
	m := curr.Mapping

	for _, obj := range objs {
		title := getString(obj, append([]string{m.TitleField}, titleFallbacks...)...)
		link := getString(obj, append([]string{m.LinkField}, linkFallbacks...)...)
		company := getString(obj, append([]string{m.CompanyField}, companyFallbacks...)...)
		location := getString(obj, append([]string{m.LocationField}, locationFallbacks...)...)
		jobTypeStr := getString(obj, append([]string{m.KindField}, kindFallbacks...)...)
		tags := getStringSlice(obj, "tags", "technologies", "skills")

		when := getEpoch(obj, "epoch", "timestamp", "time", "published", "postedAt", "date", "created_at", "published_at")
		age := time.Duration(0)
		if !when.IsZero() {
			age = now.Sub(when)
		}

		item := feed.Item{
			Source:   curr.Name,
			Title:    title,
			Link:     link,
			Company:  company,
			Location: location,

			Date: when.Local(),
			Age:  age,
		}

		if jobTypeStr != "" {
			jobType, err := feed.ParseJobType(jobTypeStr)
			if err != nil {
				log.Println("Failed to parse jobTypeStr: ", err)
			}

			item.JobType = jobType
		}

		if len(tags) > 0 {
			item.Tags = tags
		}

		out = append(out, item)
	}
	return out, nil
}

func ParseRSS(curr feed.Feed, body io.Reader) ([]feed.Item, error) {
	parser := gofeed.NewParser()

	items, err := parser.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("ERROR: parsing feed: %w", err)
	}

	out := make([]feed.Item, 0, len(items.Items))

	now := time.Now()

	for _, fi := range items.Items {
		when := time.Time{}
		if fi.PublishedParsed != nil {
			when = *fi.PublishedParsed
		} else if fi.UpdatedParsed != nil {
			when = *fi.UpdatedParsed
		}

		title := curr.Mapping.TitleField
		if fi.Title != "" {
			title = fi.Title
		}

		link := curr.Mapping.LinkField
		if fi.Link != "" {
			link = fi.Link
		}

		var age time.Duration
		if !when.IsZero() {
			age = max(now.Sub(when), 0)
		}

		jobTypeStr := fi.Custom[curr.Mapping.KindField]
		jobType, err := feed.ParseJobType(jobTypeStr)
		if err != nil {
			log.Println("Failed to parse jobTypeStr: ", err)
		}

		out = append(out, feed.Item{
			Source:   curr.Name,
			Title:    title,
			Link:     link,
			Company:  fi.Custom[curr.Mapping.CompanyField],
			Location: fi.Custom[curr.Mapping.LocationField],
			JobType:  jobType,
			Date:     when,
			Age:      age,
		})
	}

	return out, nil
}
