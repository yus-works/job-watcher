package feed

import (
	"io"
	"time"
)

type Source interface {
	GetSource() string
	GetUrl() string
	Parse(r io.Reader) ([]Item, error)
}

type BaseSource struct {
	Source string
	Url    string
}

func (b BaseSource) GetUrl() string {
	return b.Url
}

func (b BaseSource) GetSource() string {
	return b.Source
}

type Item struct {
	Source string
	Title  string
	Link   string
	Date   time.Time
	GUID   string

	Company  string
	Location string
	Kind     string // fulltime, contract, etc.
}
