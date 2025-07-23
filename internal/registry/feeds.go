package registry

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/yus-works/job-watcher/internal/feed"
	"github.com/yus-works/job-watcher/internal/parser"
)

var FEEDS = []feed.Feed{
	{
		// TODO: "https://remotive.com/api/remote-jobs?category=software-dev",
		Name: "Remotive",
		URL:  "http://localhost:8000/remotive.rss",
		Mapping: feed.ItemMap{
			CompanyField:  "company",
			LocationField: "location",
			KindField:     "type",
		},
		Parse: parser.ParseRSS,
	},
	{
		Name: "RemoteOK",
		// TODO: change to https://remoteok.com/api
		URL: "http://localhost:8000/remoteok.json",
		Mapping: feed.ItemMap{
			TitleField:    "position",
			CompanyField:  "company",
			LocationField: "location",
			KindField:     "type",
		},
		Parse: func(curr feed.Feed, body io.Reader) ([]feed.Item, error) {
			var rawItems = make([]map[string]json.RawMessage, 0)
			dec := json.NewDecoder(body)

			if err := dec.Decode(&rawItems); err != nil {
				return nil, fmt.Errorf("Failed to decode body: %w", err)
			}

			items, err := parser.ParseJSON(curr, rawItems)
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
		Mapping: feed.ItemMap{
			TitleField:    "jobTitle",
			CompanyField:  "companyName",
			LocationField: "jobGeo",
			KindField:     "jobType",
		},
		Parse: func(curr feed.Feed, body io.Reader) ([]feed.Item, error) {
			var payload = struct {
				Jobs []map[string]json.RawMessage `json:"jobs"`
			}{}

			dec := json.NewDecoder(body)
			if err := dec.Decode(&payload); err != nil {
				return nil, fmt.Errorf("Failed to decode body: %w", err)
			}

			items, err := parser.ParseJSON(curr, payload.Jobs)
			if err != nil {
				return nil, fmt.Errorf("Failed to parse: %w", err)
			}

			return items, nil
		},
	},
}
