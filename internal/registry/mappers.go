package registry

import (
	"strings"

	"github.com/yus-works/job-watcher/internal/feed"
)

type weworkMapper struct {
	feed.DefaultMapper
}

func (m weworkMapper) Title(
	decode func(selector, field string) string,
) string {
	s := decode(m.TitleField, "title")
	return strings.Split(s, ": ")[1]
}

func (m weworkMapper) Company(
	decode func(selector, field string) string,
) string {
	s := decode(m.TitleField, "title")
	return strings.Split(s, ": ")[0]
}

type infostudMapper struct {
	feed.DefaultMapper
}

func (m infostudMapper) Company(
	decode func(selector, field string) string,
) string {
	s := decode(m.DescriptionField, "title")
	return strings.Split(s, " - ")[0]
}

func (m infostudMapper) Location(
	decode func(selector, field string) string,
) string {
	s := decode(m.DescriptionField, "title")
	ssplit := strings.Split(s, " - ")
	return ssplit[len(ssplit)-1]
}

type golangprojectsMapper struct {
	feed.DefaultMapper
}

func (m golangprojectsMapper) Title(
	decode func(selector, field string) string,
) string {
	var s string
	s = decode(m.TitleField, "title")
	s = strings.Split(s, "@")[0]
	return strings.TrimSpace(s)
}

func (m golangprojectsMapper) Company(
	decode func(selector, field string) string,
) string {
	var s string
	s = decode(m.CompanyField, "company")
	p := strings.Split(s, "@")
	s = p[len(p)-1]
	return strings.TrimSpace(s)
}

//           key used             field taken from
func (m golangprojectsMapper) Location(
	decode func(selector, field string) string,
) string {
	var s string
	s = decode(m.LocationField, "location")
	s = strings.Split(s, "-")[0]
	return strings.TrimSpace(s)
}
