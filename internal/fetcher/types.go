package fetcher

import "time"

type Item struct {
	Title  string
	Body   string
	Link   string
	Date   time.Time
	GUID   string
	Source string // which feed
}
