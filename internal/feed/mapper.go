package feed

import "encoding/json"

type FieldExtractor func(obj map[string]json.RawMessage, keys ...string) string

type Mapper interface {
	Title() FieldExtractor
	Company() FieldExtractor
	Seniority() FieldExtractor
	Link() FieldExtractor
	Location() FieldExtractor
	JobType() FieldExtractor

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

func (m DefaultMapper) Title() FieldExtractor {
	return Const(m.TitleField, "title")
}

func (m DefaultMapper) Company() FieldExtractor {
	return Const(m.CompanyField, "company")
}

func (m DefaultMapper) Seniority() FieldExtractor {
	return Const(m.SeniorityField, "seniority")
}

func (m DefaultMapper) Link() FieldExtractor {
	return Const(m.LinkField, "link")
}

func (m DefaultMapper) Location() FieldExtractor {
	return Const(m.LocationField, "location")
}

func (m DefaultMapper) JobType() FieldExtractor {
	return Const(m.JobTypeField, "kind")
}
