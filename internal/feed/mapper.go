package feed

import (
	"encoding/json"
)

type FieldExtractor func(obj map[string]json.RawMessage, keys ...string) string

type jObj = map[string]json.RawMessage

// TODO: is this redundant
type Mapper interface {
	Title(getter func(val, field string) string) string
	Company(getter func(val, field string) string) string
	Seniority(getter func(val, field string) string) string
	Link(getter func(val, field string) string) string
	Location(getter func(val, field string) string) string
	JobType(getter func(val, field string) string) string

	GetConfig() Config
}

// Struct used to tell the parser what the names of the required fields are in
// each feed
//
// NOTE: Any omitted fields WILL BE IGNORED
// TODO: rename these ..Field to ..Selector (because they identify by key or whatever)
// maybe KEy
type DefaultMapper struct {
	TitleField       string
	CompanyField     string
	DescriptionField string
	SeniorityField   string
	LinkField        string
	LocationField    string
	JobTypeField     string
	DateField        string
}

type Config struct {
	DefaultMapper
}

func (m DefaultMapper) GetConfig() Config {
	return Config{m}
}

func (m DefaultMapper) Title(
	getter func(val, field string) string,
) string {
	return getter(m.TitleField, "title")
}

func (m DefaultMapper) Company(
	getter func(val, field string) string,
) string {
	return getter(m.CompanyField, "company")
}

func (m DefaultMapper) Description(
	getter func(val, field string) string,
) string {
	return getter(m.CompanyField, "company")
}

func (m DefaultMapper) Seniority(
	getter func(val, field string) string,
) string {
	return getter(m.SeniorityField, "seniority")
}

func (m DefaultMapper) Link(
	getter func(val, field string) string,
) string {
	return getter(m.LinkField, "link")
}

func (m DefaultMapper) Location(
	getter func(val, field string) string,
) string {
	return getter(m.LocationField, "location")
}

func (m DefaultMapper) JobType(
	getter func(val, field string) string,
) string {
	return getter(m.JobTypeField, "jobType")
}
