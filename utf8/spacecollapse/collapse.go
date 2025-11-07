//go:build !solution

package spacecollapse

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func CollapseSpaces(input string) string {
	var buffer strings.Builder

	countLen := 0
	inSpace := false
	for i := 0; i < len(input); {
		r, width := utf8.DecodeRuneInString(input[i:])
		i += width

		if unicode.IsSpace(r) {
			if !inSpace {
				countLen++
				inSpace = true
			}
		} else {
			countLen++
			inSpace = false
		}
	}

	buffer.Grow(countLen)
	inSpace = false

	for i := 0; i < len(input); {
		r, width := utf8.DecodeRuneInString(input[i:])
		i += width

		if unicode.IsSpace(r) {
			if !inSpace {
				buffer.WriteRune(' ')
				inSpace = true
			}
		} else {
			buffer.WriteRune(r)
			inSpace = false
		}
	}

	return buffer.String()
}
