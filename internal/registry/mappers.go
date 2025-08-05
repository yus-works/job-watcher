package registry

import (
	"strings"

	"github.com/yus-works/job-watcher/internal/feed"
)

type weworkMapper struct {
	feed.DefaultMapper
}

func (m weworkMapper) Title(
	decode func(val, field string) string,
) string {
	s := decode(m.TitleField, "title")
	return strings.Split(s, ": ")[1]
}

func (m weworkMapper) Company(
	decode func(val, field string) string,
) string {
	s := decode(m.TitleField, "title")
	return strings.Split(s, ": ")[0]
}

type infostudMapper struct {
	feed.DefaultMapper
}

func (m infostudMapper) Company(
	decode func(val, field string) string,
) string {
	s := decode(m.DescriptionField, "title")
	return strings.Split(s, " - ")[0]
}

func (m infostudMapper) Location(
	decode func(val, field string) string,
) string {
	s := decode(m.DescriptionField, "title")
	ssplit := strings.Split(s, " - ")
	return ssplit[len(ssplit)-1]
}

type golangprojectsMapper struct {
	feed.DefaultMapper
}

func (m golangprojectsMapper) Title(
	decode func(val, field string) string,
) string {
	s := decode(m.TitleField, "isthisignoredlmao?")
	return strings.Split(s, " - ")[0]
}

func (m golangprojectsMapper) Company(
	decode func(val, field string) string,
) string {
	s := decode(m.TitleField, "isthisignoredlmao?")
	ssplit := strings.Split(s, " @ ")
	return ssplit[len(ssplit)-1]
}

func (m golangprojectsMapper) Location(
	decode func(val, field string) string,
) string {
	s := decode(m.DescriptionField, "isthisignoredlmao?")
	return strings.Split(s, " - ")[0]
}
