package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/yus-works/job-watcher/internal/feed"
)

var _ feed.Mapper = feed.DefaultMapper{}

type weworkMapper struct {
	feed.DefaultMapper
}

func (m weworkMapper) Title() feed.FieldExtractor {
	return feed.Custom(
		m.TitleField, "title", func(s string) string {
			return strings.Split(s, ": ")[1]
		},
	)
}

var FEEDS = []feed.Feed{
	{
		// TODO: "https://remotive.com/api/remote-jobs?category=software-dev",
		Name: "Remotive",
		URL:  "http://localhost:8000/remotive.rss",
		Mapper: feed.DefaultMapper{
			CompanyField:  "company",
			LocationField: "location",
			JobTypeField:  "type",
		},
		Parse: feed.ParseRSS,
	},
	{
		Name: "RemoteOK",
		// TODO: change to https://remoteok.com/api
		URL: "http://localhost:8000/remoteok.json",
		Mapper: feed.DefaultMapper{
			TitleField:    "position",
			CompanyField:  "company",
			LocationField: "location",
			JobTypeField:  "type",
		},
		Parse: func(curr feed.Feed, body io.Reader) ([]feed.Item, error) {
			var rawItems = make([]map[string]json.RawMessage, 0)
			dec := json.NewDecoder(body)

			if err := dec.Decode(&rawItems); err != nil {
				return nil, fmt.Errorf("Failed to decode body: %w", err)
			}

			items, err := feed.ParseJSON(curr, rawItems)
			if err != nil {
				return nil, fmt.Errorf("Failed to parse: %w", err)
			}

			return items[1:], nil
		},
	},
	{
		Name: "Jobicy",
		// TODO: url: https://jobicy.com/api/v2/remote-jobs
		URL: "http://localhost:8000/jobicy.json",
		Mapper: feed.DefaultMapper{
			TitleField:     "jobTitle",
			CompanyField:   "companyName",
			LocationField:  "jobGeo",
			JobTypeField:   "jobType",
			DateField:      "pubDate",
			SeniorityField: "jobLevel",
		},
		Parse: func(curr feed.Feed, body io.Reader) ([]feed.Item, error) {
			var payload = struct {
				Jobs []map[string]json.RawMessage `json:"jobs"`
			}{}

			dec := json.NewDecoder(body)
			if err := dec.Decode(&payload); err != nil {
				return nil, fmt.Errorf("Failed to decode body: %w", err)
			}

			items, err := feed.ParseJSON(curr, payload.Jobs)
			if err != nil {
				return nil, fmt.Errorf("Failed to parse: %w", err)
			}

			return items, nil
		},
	},
	{
		Name: "Himalayas",
		// TODO: url: "https://himalayas.app/jobs/api"
		URL: "http://localhost:8000/himalayas.json",
		Mapper: feed.DefaultMapper{
			TitleField:     "title",
			CompanyField:   "companyName",
			LocationField:  "locationRestrictions",
			JobTypeField:   "employmentType",
			SeniorityField: "seniority",

			// NOTE: Himalayas date is last updated, not first time posted
			DateField: "pubDate",
		},
		Parse: func(curr feed.Feed, body io.Reader) ([]feed.Item, error) {
			var payload = struct {
				Jobs []map[string]json.RawMessage `json:"jobs"`
			}{}

			dec := json.NewDecoder(body)
			if err := dec.Decode(&payload); err != nil {
				return nil, fmt.Errorf("Failed to decode body: %w", err)
			}

			items, err := feed.ParseJSON(curr, payload.Jobs)
			if err != nil {
				return nil, fmt.Errorf("Failed to parse: %w", err)
			}

			return items, nil
		},
	},
	{
		Name: "WeWorkRemotely",
		// TODO: url: "https://weworkremotely.com/categories/remote-programming-jobs.rss"
		URL: "http://localhost:8000/remote-programming-jobs.rss",
		Mapper: weworkMapper{
			DefaultMapper: feed.DefaultMapper{
				TitleField:    "category",
				CompanyField:  "title",
				LocationField: "region",
				DateField:     "pubDate",
			},
		},
		Parse: feed.ParseRSS,
	},
}
