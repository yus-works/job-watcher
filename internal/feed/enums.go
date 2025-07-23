package feed

import (
	"fmt"
	"regexp"
)

var ignoreNonLetters = regexp.MustCompile(`[^A-Za-z]+`)

type JobType string

const (
	UnknownJobType JobType = ""
	FullTime       JobType = "fulltime"
	PartTime       JobType = "parttime"
	Contract       JobType = "contract"
	Internship     JobType = "internship"
)

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
		return UnknownJobType, fmt.Errorf("Failed to parse (%s)", s)
	}
}

type Seniority string

const (
	UnknownSeniority           = ""
	Intern           Seniority = "intern"
	Junior           Seniority = "junior"
	Medior           Seniority = "medior"
	Senior           Seniority = "senior"
)

// ParseSeniority normalizes s (drops non‑letters) and returns the matching Seniority.
func ParseSeniority(s string) (Seniority, error) {
	seniority := ignoreNonLetters.ReplaceAllString(s, "")

	switch seniority {
	case "intern":
		return Intern, nil
	case "junior":
		return Junior, nil
	case "medior", "intermediate", "midweight":
		return Medior, nil
	case "senior":
		return Senior, nil
	default:
		return UnknownSeniority, fmt.Errorf("Failed to parse (%s)", s)
	}
}
