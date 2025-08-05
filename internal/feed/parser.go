package feed

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
)

func makeGetterJSON(
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

func makeGetterRSS(
	log *logrus.Entry,
	feeditem FeedItemWrapped,
	curr Feed,
) func(val, field string) string {
	return func(selector, field string) string {
		if v, err := feeditem.Get(selector); err == nil {
			return v
		}
		log.WithFields(logrus.Fields{
			"field": field,
			"with":  selector,
			"for":   curr.Name,
		}).Warn("get")
		return ""
	}
}

func ParseJSON(log *logrus.Entry, curr Feed, objs []map[string]json.RawMessage) ([]JobItem, error) {
	out := make([]JobItem, 0, len(objs))
	now := time.Now()
	m := curr.Mapper

	for _, obj := range objs {
		objCopy := obj

		getter := makeGetterJSON(objCopy)
		title := m.Title(getter)
		link := m.Link(getter)
		company := m.Company(getter)
		location := m.Location(getter)
		seniorityStr := m.Seniority(getter)
		jobTypeStr := m.JobType(getter)

		tags := getStringSlice(obj, "tags", "technologies", "skills")

		when := getEpoch(obj, append([]string{m.GetConfig().DateField}, epochFallbacks...)...)
		age := time.Duration(0)
		if !when.IsZero() {
			age = now.Sub(when)
		}

		item := JobItem{
			Source:   curr.Name,
			Title:    title,
			Link:     link,
			Company:  company,
			Location: location,

			Date: when.Local(),
			Age:  age,
		}

		if jobTypeStr != "" {
			jobType, err := ParseJobType(log, jobTypeStr)
			if err != nil {
				log.WithError(err).Warn("parse jobTypeStr")
			}

			item.JobType = jobType
		}

		if seniorityStr != "" {
			seniority, err := ParseSeniority(log, seniorityStr)
			if err != nil {
				log.WithError(err).Warn("parse seniorityStr")
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

func ParseRSS(log *logrus.Entry, curr Feed, body io.Reader) ([]JobItem, error) {
	parser := gofeed.NewParser()

	items, err := parser.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("ERROR: parsing feed: %w", err)
	}

	out := make([]JobItem, 0, len(items.Items))

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

		getter := makeGetterRSS(log, feeditem, curr)

		title := curr.Mapper.Title(getter)
		company := curr.Mapper.Company(getter)
		location := curr.Mapper.Location(getter)

		var age time.Duration
		if !when.IsZero() {
			age = max(now.Sub(when), 0)
		}

		item := JobItem{
			Source:   curr.Name,
			Link:     getter(cfg.LinkField, "link"),
			Title:    title,
			Company:  company,
			Location: location,
			Date:     when,
			Age:      age,
		}

		if cfg.JobTypeField != "" {
			jobTypeStr := getter(cfg.JobTypeField, "job type")
			if jobTypeStr != "" {
				jobType, err := ParseJobType(log, jobTypeStr)
				if err != nil {
					log.WithError(err).Warn("parse jobTypeStr")
				}

				item.JobType = jobType
			}
		}

		out = append(out, item)
	}

	return out, nil
}
