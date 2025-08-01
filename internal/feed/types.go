package feed

import (
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

type JobItem struct {
	Hash     int64
	Source   string
	Title    string
	Link     string
	Company  string
	Location string

	Seniority Seniority
	JobType   JobType
	Date      time.Time

	Age        time.Duration
	InsertedAt time.Time
	Score      float64

	// TODO: some kind of tag enum/normalization
	Tags []string
}

type Feed struct {
	Name   string
	URL    string
	Mapper Mapper
	Parse  func(
		log *logrus.Entry,
		curr Feed,
		body io.Reader,
	) ([]JobItem, error)
}
