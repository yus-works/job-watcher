package jobs

import "github.com/yus-works/job-watcher/internal/feed"

type DisplayItem struct {
	feed.Item
	JobType string
	Date    string
	Age     string
}
