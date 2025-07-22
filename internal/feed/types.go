package feed

import (
	"fmt"
	"io"
	"regexp"
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
}

type Feed struct {
	Name    string
	URL     string
	Mapping ItemMap
	Parse   func(curr Feed, body io.Reader) ([]Item, error)
}

type JobType string

const (
	Unknown    JobType = ""
	FullTime   JobType = "fulltime"
	PartTime   JobType = "parttime"
	Contract   JobType = "contract"
	Internship JobType = "internship"
)

var ignoreNonLetters = regexp.MustCompile(`[^A-Za-z]+`)

// ParseJobType normalizes s (drops non‑letters) and returns the matching JobType.
func ParseJobType(s string) (JobType, error) {
	jobType := ignoreNonLetters.ReplaceAllString(s, "")

	switch jobType {
	case "fulltime":
		return FullTime, nil
	case "parttime":
		return PartTime, nil
	case "contract":
		return Contract, nil
	case "internship":
		return Internship, nil
	default:
		return Unknown, fmt.Errorf("Failed to parse (%s)", s)
	}
}

type Item struct {
	Source   string
	Title    string
	Link     string
	Company  string
	Location string

	JobType JobType
	Date    time.Time
	Age     time.Duration

	// TODO: some kind of tag enum/normalization
	Tags []string
}

type DisplayItem struct {
	Item
	JobType string
	Date    string
	Age     string
}
