package store

import (
	"database/sql"
	"fmt"
	"hash/fnv"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/sirupsen/logrus"
	"github.com/yus-works/job-watcher/internal/feed"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

func normalize(s string) string {
	s = strings.TrimSpace(s)
	s = cases.Fold().String(s)
	t := norm.NFKD.String(s)
	buf := make([]rune, 0, len(t))
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		buf = append(buf, r)
	}
	return strings.Join(strings.Fields(string(buf)), " ")
}

func HashNormalized64(s string) int64 {
	h := fnv.New64a()
	h.Write([]byte(normalize(s)))
	return int64(h.Sum64() & math.MaxInt64) // zero out MSB
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func NormalizeTag(s string) string {
	s = strings.ToLower(s)
	s = norm.NFKD.String(s)
	s = nonAlnum.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func tolerantParse[T feed.Seniority | feed.JobType](
	log *logrus.Entry,
	s sql.NullString,
	parser func(log *logrus.Entry, s string) (T, error),
) T {
	var enum T
	var err error

	if s.Valid && s.String != "" {
		enum, err = parser(log, s.String)
		if err != nil {
			log.WithError(err).Warn("parse")
		}
	}

	return enum
}

func parseDate(
	log *logrus.Entry,
	dateStr sql.NullString,
) time.Time {
	if !dateStr.Valid && dateStr.String == "" {
		err := fmt.Errorf("invalid or missing date in DB: %q", dateStr.String)
		log.WithError(err).Error("parse date")
		return time.Time{}
	}

	if t, err := time.Parse(time.RFC3339, dateStr.String); err == nil {
		return t
	} else if t, err := time.Parse("2006-01-02", dateStr.String); err == nil {
		return t
	} else {
		err = fmt.Errorf("invalid date in DB: %q: %w", dateStr.String, err)
		log.WithError(err).Error("parse date")
		return time.Time{}
	}
}
