package fetcher

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"time"
)

func decodeItems(r io.Reader) ([]Item, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	trim := bytes.TrimSpace(data)
	if len(trim) == 0 {
		return nil, errors.New("empty body")
	}

	switch trim[0] {
	case '{', '[':
		return decodeJSONItems(trim)
	case '<':
		return decodeXMLItems(trim)
	default:
		return nil, errors.New("unknown format (not JSON or XML)")
	}
}

func decodeJSONItems(b []byte) ([]Item, error) {
	var arr []struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}

	if err := json.Unmarshal(b, &arr); err == nil {
		return mapArray(arr), nil
	}
	return nil, errors.New("JSON shape not recognised")
}

func mapArray(in []struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}) []Item {
	out := make([]Item, 0, len(in))
	for _, j := range in {
		var runes []rune

		runes = []rune(j.Title)
		title := string(runes[:5])

		runes = []rune(j.Body)
		body := string(runes[:10])

		out = append(out, Item{
			Title: title,
			Body:  body,
		})
	}
	return out
}

func decodeXMLItems(b []byte) ([]Item, error) {
	// minimal RSS struct
	var rss struct {
		Channel struct {
			Items []struct {
				Title string `xml:"title"`
				Body  string `xml:"body"`
			} `xml:"item"`
		} `xml:"channel"`
	}

	if err := xml.Unmarshal(b, &rss); err != nil {
		return nil, err
	}

	out := make([]Item, 0, len(rss.Channel.Items))
	for _, it := range rss.Channel.Items {
		out = append(out, Item{
			Title: strings.TrimSpace(it.Title),
			Body:  strings.TrimSpace(it.Body),
		})
	}
	return out, nil
}

func parseDate(s string) time.Time {
	// try RFC1123 (first for RSS), then ISO8601 & fall back to zero time.
	formats := []string{
		time.RFC1123Z, // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,  // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC3339,  // "2006-01-02T15:04:05Z07:00"
		"2006-01-02",  // plain date
	}
	for _, f := range formats {
		if t, err := time.Parse(f, strings.TrimSpace(s)); err == nil {
			return t
		}
	}
	return time.Time{}
}
