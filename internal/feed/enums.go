package feed

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

var ignoreNonLetters = regexp.MustCompile(`[^A-Za-z]+`)

func normalize(s string) string {
	return strings.ToLower(ignoreNonLetters.ReplaceAllString(s, ""))
}

type JobType string

const (
	UnknownJobType JobType = ""
	FullTime       JobType = "fulltime"
	PartTime       JobType = "parttime"
	Contract       JobType = "contract"
	Internship     JobType = "internship"
)

var JOB_TYPES = []string{"fulltime", "parttime", "contract", "internship"}

// ParseJobType normalizes s (drops non‑letters) and returns the matching JobType.
func ParseJobType(log *logrus.Entry, s string) (JobType, error) {
	jobType := normalize(s)

	switch jobType {
	case "fulltime":
		return FullTime, nil
	case "parttime":
		return PartTime, nil
	case "contract", "contractor":
		return Contract, nil
	case "internship":
		return Internship, nil
	default:
		log.WithFields(logrus.Fields{
			"jobType": s,
		}).Warn("parse")
		return UnknownJobType, fmt.Errorf("Failed to parse jobType (%s)", s)
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

var SENIORITIES = []string{"intern", "junior", "medior", "senior"}

// ParseSeniority normalizes s (drops non‑letters) and returns the matching Seniority.
func ParseSeniority(log *logrus.Entry, s string) (Seniority, error) {
	seniority := normalize(s)

	switch seniority {
	case "intern":
		return Intern, nil
	case "junior", "entrylevel", "entryleveljunior":
		return Junior, nil
	case "medior", "intermediate", "midweight":
		return Medior, nil
	case "senior":
		return Senior, nil
	case "any":
		return UnknownSeniority, nil
	default:
		log.WithFields(logrus.Fields{
			"seniority": s,
		}).Warn("parse")
		return UnknownSeniority, fmt.Errorf("Failed to parse seniority(%s)", s)
	}
}
