package feed

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/mmcdole/gofeed"
)

func makeExtractor(
	o map[string]json.RawMessage,
) func(val, field string) string {
	// returns a closure that has o baked in
	return func(val, field string) string {
		return getString(
			o,
			append([]string{val}, getFallbacks(field)...)...,
		)
	}
}

func ParseJSON(curr Feed, objs []map[string]json.RawMessage) ([]Item, error) {
	out := make([]Item, 0, len(objs))
	now := time.Now()
	m := curr.Mapper

	for _, obj := range objs {
		objCopy := obj

		extractor := makeExtractor(objCopy)
		title := m.Title(extractor)
		link := m.Link(extractor)
		company := m.Company(extractor)
		location := m.Location(extractor)
		seniorityStr := m.Seniority(extractor)
		jobTypeStr := m.JobType(extractor)

		tags := getStringSlice(obj, "tags", "technologies", "skills")

		when := getEpoch(obj, append([]string{m.GetConfig().DateField}, epochFallbacks...)...)
		age := time.Duration(0)
		if !when.IsZero() {
			age = now.Sub(when)
		}

		item := Item{
			Source:   curr.Name,
			Title:    title,
			Link:     link,
			Company:  company,
			Location: location,

			Date: when.Local(),
			Age:  age,
		}

		if jobTypeStr != "" {
			jobType, err := ParseJobType(jobTypeStr)
			if err != nil {
				log.Println("Failed to parse jobTypeStr: ", err)
			}

			item.JobType = jobType
		}

		if seniorityStr != "" {
			seniority, err := ParseSeniority(seniorityStr)
			if err != nil {
				log.Println("Failed to parse seniorityStr: ", err)
			}

			item.Seniority = seniority
		}

		if len(tags) > 0 {
			item.Tags = tags
		}

		out = append(out, item)
	}
	return out, nil
}

func ParseRSS(curr Feed, body io.Reader) ([]Item, error) {
	parser := gofeed.NewParser()

	items, err := parser.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("ERROR: parsing feed: %w", err)
	}

	out := make([]Item, 0, len(items.Items))

	now := time.Now()

	for _, fi := range items.Items {
		when := time.Time{}
		if fi.PublishedParsed != nil {
			when = *fi.PublishedParsed
		} else if fi.UpdatedParsed != nil {
			when = *fi.UpdatedParsed
		}

		cfg := curr.Mapper.GetConfig()
		feeditem := FeedItemWrapped{fi}

		extractor := func(selector, field string) string {
			if v, err := feeditem.Get(selector); err == nil {
				return v
			}
			log.Printf(
				"Failed to get %s with (%s) for (%s)",
				field, selector, curr.Name,
			)
			return ""
		}

		title := curr.Mapper.Title(func(_, _ string) string {
			return extractor(cfg.TitleField, "title")
		})
		company := curr.Mapper.Company(func(_, _ string) string {
			return extractor(cfg.CompanyField, "company")
		})

		var age time.Duration
		if !when.IsZero() {
			age = max(now.Sub(when), 0)
		}

		item := Item{
			Source:   curr.Name,
			Link:     extractor(cfg.LinkField, "link"),
			Title:    title,
			Company:  company,
			Location: extractor(cfg.LocationField, "location"),
			Date:     when,
			Age:      age,
		}

		if cfg.JobTypeField != "" {
			jobTypeStr := extractor(cfg.JobTypeField, "job type")
			if jobTypeStr != "" {
				jobType, err := ParseJobType(jobTypeStr)
				if err != nil {
					log.Println("Failed to parse jobTypeStr: ", err)
				}

				item.JobType = jobType
			}
		}

		out = append(out, item)
	}

	return out, nil
}
