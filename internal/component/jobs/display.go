package jobs

import (
	"github.com/yus-works/job-watcher/internal/feed"
)

type DisplayItem struct {
	feed.JobItem
	Seniority string
	JobType   string
	Date      string
	Age       string
}

func NewDisplayItem(i feed.JobItem) DisplayItem {
	return DisplayItem{
		JobItem: i,

		Seniority: string(i.Seniority),
		JobType:   string(i.JobType),
		Date:      i.Date.Local().Format("2006-01-02"),
		Age:       feed.HumanAgeGreedy(i.Age),
	}
}
