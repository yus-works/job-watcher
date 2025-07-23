package feed

import (
	"io"
	"time"
)

// Struct used to tell the parser what the names of the required fields are in
// each feed
//
// Common fields like Title and URL can usually be omitted because the parser
// can usually pick those up automatically
type ItemMap struct {
	TitleField    string
	LinkField     string
	CompanyField  string
	LocationField string
	KindField     string
	DateField     string
}

type Feed struct {
	Name    string
	URL     string
	Mapping ItemMap
	Parse   func(curr Feed, body io.Reader) ([]Item, error)
}

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
