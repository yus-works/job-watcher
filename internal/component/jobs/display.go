package jobs

import (
	"github.com/yus-works/job-watcher/internal/feed"
	"github.com/yus-works/job-watcher/internal/parser"
)

type DisplayItem struct {
	feed.Item
	JobType string
	Date    string
	Age     string
}

func NewDisplayItem(i feed.Item) DisplayItem {
	return DisplayItem{
		Item: i,

		JobType: string(i.JobType),
		Date:    i.Date.Local().Format("2006-01-02"),
		Age:     parser.HumanAgeGreedy(i.Age),
	}
}
