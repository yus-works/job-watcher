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
	name    string
	url     string
	mapping ItemMap
	parser  func(curr Feed, body io.Reader) ([]Item, error)
}

func (f Feed) Name() string {
	return f.name
}
func (f Feed) URL() string {
	return f.url
}
func (f Feed) Mapping() ItemMap {
	return f.mapping
}
func (f Feed) Parse(body io.Reader) ([]Item, error) {
	return f.parser(f, body)
}

func NewFeed(
	name, url string,
	mapping ItemMap,
	parser func(curr Feed, body io.Reader) ([]Item, error),
) Feed {
	return Feed{
		name:    name,
		url:     url,
		mapping: mapping,
		parser:  parser,
	}
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

	JobType    JobType
	JobTypeStr string

	// TODO: some kind of tag enum/normalization
	Tags []string

	Date   time.Time
	Age    time.Duration
	AgeStr string
}
