package speller

import (
	"strings"
	"unicode/utf8"
)

func Reverse(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := len(s); i > 0; {
		r, size := utf8.DecodeLastRuneInString(s[:i])
		i -= size
		b.WriteRune(r)
	}

	return b.String()
}
