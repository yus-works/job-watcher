package feed

import (
	"io"
	"time"
)

type Feed interface {
	GetSource() string
	GetUrl() string
	Parse(r io.Reader) ([]Item, error)
}

type Item struct {
	Source string
    Title    string
    Link     string
    Date     time.Time
    GUID     string

    Company  string
    Location string
    Kind     string // fulltime, contract, etc.
}
