package jobs

import (
	"html"
	"strings"
	"unicode"

	"github.com/yus-works/job-watcher/internal/feed"
)

func cleanText(s string) string {
	s = strings.TrimSpace(s)
	// bounded multi-pass for double-encoded inputs like &amp;amp;
	for range 3 {
		u := html.UnescapeString(s)
		if u == s {
			break
		}
		s = u
	}
	s = strings.ReplaceAll(s, "\u00A0", " ")
	return strings.TrimFunc(s, unicode.IsSpace)
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
	di.Company = cleanText(di.Company)

	return di
}
