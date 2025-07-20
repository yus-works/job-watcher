package feed

import "time"

// Struct used to tell the parser what the names of the required fields are in
// each feed
//
// Common fields like Title and URL can usually be omitted because the parser
// can usually pick those up automatically
type ItemMap struct {
	Title    string
	Link     string
	Company  string
	Location string
	Kind     string // fulltime, contract, etc.
}

type Feed struct {
	Name    string
	URL     string
	Mapping ItemMap
}

type Item struct {
	Source   string
	Title    string
	Link     string
	Company  string
	Location string
	Kind     string // fulltime, contract, etc.
	Date     time.Time
}
