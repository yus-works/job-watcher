package feed

import (
	"io"
	"time"
)

type Item struct {
	Source   string
	Title    string
	Link     string
	Company  string
	Location string

	Seniority Seniority
	JobType   JobType
	Date      time.Time
	Age       time.Duration

	// TODO: some kind of tag enum/normalization
	Tags []string
}

type Feed struct {
	Name   string
	URL    string
	Mapper Mapper
	Parse  func(curr Feed, body io.Reader) ([]Item, error)
}
