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
		Name: "Remotive",
		URL:  "http://localhost:8000/remotive.rss",
		Mapping: feed.ItemMap{
			CompanyField:  "company",
			LocationField: "location",
			KindField:     "type",
		},
		Parse: parser.ParseRSS,
	},

	/*
		https://remoteok.com/api
		type: JSON
		structure:
		- list of objects
		- first object is info
		relevant fields:
		- epoch
		- company
		- position
		- tags
		- location
		- url
		info: cant find
	*/
	{
		Name: "RemoteOK",
		URL:  "http://localhost:8000/remoteok.json",
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
}
