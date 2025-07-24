package feed

import "encoding/json"

type FieldExtractor func(obj map[string]json.RawMessage, keys ...string) string

type jObj = map[string]json.RawMessage

type Mapper interface {
	Title(decode func(val, field string) string) string
	Company(decode func(val, field string) string) string
	Seniority(decode func(val, field string) string) string
	Link(decode func(val, field string) string) string
	Location(decode func(val, field string) string) string
	JobType(decode func(val, field string) string) string

	GetConfig() Config
}

// Struct used to tell the parser what the names of the required fields are in
// each feed
//
// Common fields like Title and URL can usually be omitted because the parser
// can usually pick those up automatically
type DefaultMapper struct {
	TitleField     string
	CompanyField   string
	SeniorityField string
	LinkField      string
	LocationField  string
	JobTypeField   string
	DateField      string
}

type Config struct {
	DefaultMapper
}

func (m DefaultMapper) GetConfig() Config {
	return Config{m}
}

func (m DefaultMapper) Title(
	decode func(val, field string) string,
) string {
	return decode(m.TitleField, "title")
}

func (m DefaultMapper) Company(
	decode func(val, field string) string,
) string {
	return decode(m.CompanyField, "company")
}

func (m DefaultMapper) Seniority(
	decode func(val, field string) string,
) string {
	return decode(m.SeniorityField, "seniority")
}

func (m DefaultMapper) Link(
	decode func(val, field string) string,
) string {
	return decode(m.LinkField, "link")
}

func (m DefaultMapper) Location(
	decode func(val, field string) string,
) string {
	return decode(m.LocationField, "location")
}

func (m DefaultMapper) JobType(
	decode func(val, field string) string,
) string {
	return decode(m.JobTypeField, "jobType")
}
