package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yus-works/job-watcher/internal/feed"
)

var _ feed.Mapper = feed.DefaultMapper{}

type weworkMapper struct {
	feed.DefaultMapper
}

func (m weworkMapper) Title(
	decode func(val, field string) string,
) string {
	s := decode(m.TitleField, "title")
	return strings.Split(s, ": ")[1]
}

func (m weworkMapper) Company(
	decode func(val, field string) string,
) string {
	s := decode(m.TitleField, "title")
	return strings.Split(s, ": ")[0]
}

type infostudMapper struct {
	feed.DefaultMapper
}

func (m infostudMapper) Company(
	decode func(val, field string) string,
) string {
	s := decode(m.DescriptionField, "title")
	return strings.Split(s, " - ")[0]
}

func (m infostudMapper) Location(
	decode func(val, field string) string,
) string {
	s := decode(m.DescriptionField, "title")
	ssplit := strings.Split(s, " - ")
	return ssplit[len(ssplit)-1]
}

// TODO: add refreshPeriod field to each feed
// that triggers the refresh routine or something
var FEEDS = []feed.Feed{
	{
		Name: "Remotive",
		URL:  "https://remotive.com/api/remote-jobs?category=software-dev",
		Mapper: feed.DefaultMapper{
			TitleField:    "title",
			LinkField:     "link",
			CompanyField:  "company",
			LocationField: "location",
			JobTypeField:  "type",
		},
		Parse: feed.ParseRSS,
	},
	{
		Name: "RemoteOK",
		URL:  "https://remoteok.com/api",
		Mapper: feed.DefaultMapper{
			TitleField:    "position",
			CompanyField:  "company",
			LocationField: "location",
			JobTypeField:  "type",
		},
		Parse: func(log *logrus.Entry, curr feed.Feed, body io.Reader) ([]feed.JobItem, error) {
			var rawItems = make([]map[string]json.RawMessage, 0)
			dec := json.NewDecoder(body)

			if err := dec.Decode(&rawItems); err != nil {
				return nil, fmt.Errorf("Failed to decode body: %w", err)
			}

			items, err := feed.ParseJSON(log, curr, rawItems)
			if err != nil {
				return nil, fmt.Errorf("Failed to parse: %w", err)
			}

			return items[1:], nil
		},
	},
	{
		Name: "Jobicy",
		URL:  "https://jobicy.com/api/v2/remote-jobs",
		Mapper: feed.DefaultMapper{
			TitleField:     "jobTitle",
			CompanyField:   "companyName",
			LocationField:  "jobGeo",
			JobTypeField:   "jobType",
			DateField:      "pubDate",
			SeniorityField: "jobLevel",
		},
		Parse: func(log *logrus.Entry, curr feed.Feed, body io.Reader) ([]feed.JobItem, error) {
			var payload = struct {
				Jobs []map[string]json.RawMessage `json:"jobs"`
			}{}

			dec := json.NewDecoder(body)
			if err := dec.Decode(&payload); err != nil {
				return nil, fmt.Errorf("Failed to decode body: %w", err)
			}

			items, err := feed.ParseJSON(log, curr, payload.Jobs)
			if err != nil {
				return nil, fmt.Errorf("Failed to parse: %w", err)
			}

			return items, nil
		},
	},
	{
		Name: "Himalayas",
		URL:  "https://himalayas.app/jobs/api",
		Mapper: feed.DefaultMapper{
			TitleField:     "title",
			CompanyField:   "companyName",
			LocationField:  "locationRestrictions",
			JobTypeField:   "employmentType",
			SeniorityField: "seniority",

			// NOTE: Himalayas date is last updated, not first time posted
			DateField: "pubDate",
		},
		Parse: func(log *logrus.Entry, curr feed.Feed, body io.Reader) ([]feed.JobItem, error) {
			var payload = struct {
				Jobs []map[string]json.RawMessage `json:"jobs"`
			}{}

			dec := json.NewDecoder(body)
			if err := dec.Decode(&payload); err != nil {
				return nil, fmt.Errorf("Failed to decode body: %w", err)
			}

			items, err := feed.ParseJSON(log, curr, payload.Jobs)
			if err != nil {
				return nil, fmt.Errorf("Failed to parse: %w", err)
			}

			return items, nil
		},
	},
	{
		Name: "WeWorkRemotely",
		URL:  "https://weworkremotely.com/categories/remote-programming-jobs.rss",
		Mapper: weworkMapper{
			DefaultMapper: feed.DefaultMapper{
				// both will be post processed by custom Title()
				TitleField:   "title",
				LinkField:    "link",
				CompanyField: "title",

				LocationField: "region",
				DateField:     "pubDate",
			},
		},
		Parse: feed.ParseRSS,
	},
	{
		Name: "Infostud",
		URL:  "http://rss.infostud.com/poslovi/",
		Mapper: infostudMapper{
			DefaultMapper: feed.DefaultMapper{
				TitleField:    "title",
				LinkField:     "link",
				CompanyField:  "description",
				LocationField: "description",
				DateField:     "pubDate",
			},
		},
		Parse: feed.ParseRSS,
	},
}
