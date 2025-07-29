package feed

import (
	"io"
	"time"
)

// TODO: wait is Item and Job totally redundant? could I just use one of them?
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

func Identifier()
j.Title + "|" + j.Company

type Feed struct {
	Name   string
	URL    string
	Mapper Mapper
	Parse  func(curr Feed, body io.Reader) ([]Item, error)
}
