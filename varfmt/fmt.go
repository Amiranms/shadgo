//go:build !solution

package varfmt

// package main

import (
	"fmt"
	"strconv"
	"strings"
)

func Sprintf(format string, args ...interface{}) string {

	// memory allocation
	var builder strings.Builder
	builder.Grow(len(format) + len(args))

	argsStr := make([]string, len(args))

	for i := range len(args) {
		argsStr[i] = fmt.Sprint(args[i])
	}
	currentPlaceholderIdx := 0

	for i := 0; i < len(format); {
		if format[i] == '{' {
			openBracketPos := i
			closeBracketPos := strings.IndexByte(format[openBracketPos:], '}')

			if closeBracketPos == -1 {
				builder.WriteString(format[openBracketPos:])
				break
			}
			closeBracketPos += openBracketPos

			idxStr := format[openBracketPos+1 : closeBracketPos]
			var argIdx int

			if idxStr == "" {
				argIdx = currentPlaceholderIdx
			} else {
				idxInt, err := strconv.Atoi(idxStr)
				if err != nil {
					builder.WriteString(format[openBracketPos : closeBracketPos+1])
					i = closeBracketPos + 1
					currentPlaceholderIdx++
					continue
				}
				argIdx = idxInt
			}

			currentPlaceholderIdx++

			if argIdx >= 0 && argIdx < len(argsStr) {
				builder.WriteString(argsStr[argIdx])
			}

			i = closeBracketPos + 1

		} else {
			builder.WriteByte(format[i])
			i++
		}

	}

	return builder.String()

}
