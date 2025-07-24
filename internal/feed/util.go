package feed

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func HumanAgeGreedy(dur time.Duration) string {
	if dur <= 0 {
		return "0h"
	}

	const (
		hour       = time.Hour
		dayHours   = 24 * hour
		weekHours  = 7 * dayHours
		monthHours = 30 * dayHours
	)

	months := dur / monthHours
	dur -= months * monthHours

	weeks := dur / weekHours
	dur -= weeks * weekHours

	days := dur / dayHours
	dur -= days * dayHours

	hours := dur / hour

	parts := make([]string, 0, 4)
	if months > 0 {
		parts = append(parts, fmt.Sprintf("%dmo", months))
	}
	if weeks > 0 {
		parts = append(parts, fmt.Sprintf("%dw", weeks))
	}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}

	if len(parts) == 0 {
		return "0h"
	}
	return strings.Join(parts, " ")
}

// return first value that is accessible by a key from keys
func firstRaw(obj map[string]json.RawMessage, keys ...string) (json.RawMessage, bool) {
	for _, k := range keys {
		if k == "" {
			continue
		}
		if v, ok := obj[k]; ok {
			return v, true
		}
	}
	return nil, false
}

func getString(obj map[string]json.RawMessage, keys ...string) string {
	if b, ok := firstRaw(obj, keys...); ok {
		var s string
		if json.Unmarshal(b, &s) == nil {
			return s
		}
	}
	return ""
}

func getStringSlice(obj map[string]json.RawMessage, keys ...string) []string {
	if b, ok := firstRaw(obj, keys...); ok {
		var ss []string
		if json.Unmarshal(b, &ss) == nil {
			return ss
		}
	}
	return nil
}

func getEpoch(obj map[string]json.RawMessage, keys ...string) time.Time {
	if b, ok := firstRaw(obj, keys...); ok {
		// number?
		var n int64
		if json.Unmarshal(b, &n) == nil {
			return time.Unix(n, 0).UTC()
		}
		// string?
		var s string
		if json.Unmarshal(b, &s) == nil {
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				return time.Unix(n, 0).UTC()
			}
			// try RFC3339
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				return t.UTC()
			}
		}
	}
	return time.Time{}
}

func getFallbacks(field string) []string {
	switch field {
	case "title":
		return titleFallbacks
	case "link":
		return linkFallbacks
	case "company":
		return companyFallbacks
	case "location":
		return locationFallbacks
	// TODO: rename kind
	case "kind":
		return kindFallbacks
	case "seniority":
		return seniorityFallbacks
	default:
		return []string{}
	}
}

// Const(val string) returns a callback that just returns val
// This is a helper to avoid writing functions for the callback defined fields
// in ItemMap when there's no need for custom logic and we just want a string
// Usage:
//
//	feed.ItemMap {
//	  TitleField: Const("jobTitle", "title"),
//	  CompanyField: Const("companyName", "company"),
//	...
//	}
func Const(val, field string) func(obj map[string]json.RawMessage, keys ...string) string {
	return func(obj map[string]json.RawMessage, keys ...string) string {
		return getString(obj, append([]string{val}, getFallbacks(field)...)...)
	}
}

func Custom(val, field string, postProcessor func(s string) string) func(obj map[string]json.RawMessage, keys ...string) string {
	return func(obj map[string]json.RawMessage, keys ...string) string {
		str := getString(obj, append([]string{val}, getFallbacks(field)...)...)
		return postProcessor(str)
	}
}

func GetEpoch(val, field string) func(obj map[string]json.RawMessage, keys ...string) time.Time {
	return func(obj map[string]json.RawMessage, keys ...string) time.Time {
		return getEpoch(obj, append([]string{val}, epochFallbacks...)...)
	}
}
