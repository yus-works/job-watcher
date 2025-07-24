package feed

import (
	"fmt"

	"github.com/mmcdole/gofeed"
)

type FeedItemWrapped struct{ *gofeed.Item }

func (i FeedItemWrapped) Get(field string) (string, error) {
	str := ""

	switch field {
	case "title":
		str = i.Title
	case "link":
		str = i.Link
	default:
		str = i.Custom[field]
	}

	if str == "" {
		return "", fmt.Errorf("ERROR: Couldn't find value of field (%s): ", field)
	}

	return str, nil
}
