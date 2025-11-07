//go:build !solution

package main

import (
	"fmt"
	"os"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func mapRead(paths []string, fn func(string) ([]byte, error)) []string {
	result := make([]string, len(paths))
	for i, s := range paths {
		bytes, e := os.ReadFile(s)
		check(e)
		result[i] = string(bytes)
	}
	return result
}

func readFromFiles(paths []string) string {
	return strings.Join(mapRead(paths, os.ReadFile), "\n")
}

func count(content string) map[string]int {
	counter := make(map[string]int)
	for _, s := range strings.Split(content, "\n") {
		counter[s]++
	}
	return counter
}

func wordcount(paths []string) map[string]int {
	return count(readFromFiles(paths))
}

func prettyMapPrint(m map[string]int, delim string, pred func(int) bool) {

	for k, v := range m {
		if pred(v) {
			fmt.Printf("%d%s%s\n", v, delim, k)
		}
	}
}

func main() {
	var paths = os.Args[1:]
	prettyMapPrint(wordcount(paths), "\t", func(c int) bool {
		return c >= 2
	})
}
