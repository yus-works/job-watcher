package jobs

import (
	"html"
	"strings"

	"github.com/yus-works/job-watcher/internal/feed"
)

func cleanText(s string) string {
	return html.UnescapeString(strings.TrimSpace(s))
}

type DisplayItem struct {
	feed.JobItem
	Seniority string
	JobType   string
	Date      string
	Age       string
}

func NewDisplayItem(i feed.JobItem) DisplayItem {
	di := DisplayItem{
		JobItem: i,

		Seniority: string(i.Seniority),
		JobType:   string(i.JobType),
		Date:      i.Date.Local().Format("2006-01-02"),
		Age:       feed.HumanAgeGreedy(i.Age),
	}

	di.Title = cleanText(di.Title)
	di.Title = cleanText(di.Title)
	di.Company = cleanText(di.Company)
	di.Company = cleanText(di.Company)

	return di
}
