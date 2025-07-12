package feed

import (
	"io"
	"time"
)

type Feed interface {
	GetUrl() string
	Parse(r io.Reader) ([]Item, error)
}

type Item struct {
    Title    string
    Link     string
    Date     time.Time
    GUID     string

    Company  string
    Location string
    Kind     string // fulltime, contract, etc.
}
