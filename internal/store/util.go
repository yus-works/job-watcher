package store

import (
	"hash/fnv"
	"math"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

func Normalize(s string) string {
	s = strings.TrimSpace(s)
	s = cases.Fold().String(s)
	t := norm.NFKD.String(s)
	buf := make([]rune, 0, len(t))
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		buf = append(buf, r)
	}
	return strings.Join(strings.Fields(string(buf)), " ")
}

func HashNormalized64(s string) int64 {
	h := fnv.New64a()
	h.Write([]byte(Normalize(s)))
	return int64(h.Sum64() & math.MaxInt64) // zero out MSB
}
