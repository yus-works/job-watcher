package feed

import (
	"fmt"
	"regexp"
)

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
