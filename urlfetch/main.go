//go:build !solution

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func getPageContent(rawURL string) string {
	url, err := url.Parse(rawURL)
	check(err)

	res, err := http.Get(url.String())
	check(err)

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if res.StatusCode < 200 && res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		os.Exit(res.StatusCode)
	}
	check(err)
	return string(body)
}

func mapRead(urls []string, f func(url string) string) string {
	var builder strings.Builder
	for _, url := range urls {
		builder.WriteString(f(url))
	}
	return builder.String()
}

func readUrls(urls []string) string {
	return mapRead(urls, getPageContent)
}

func main() {
	urls := os.Args[1:]

	fmt.Println(readUrls(urls))
}
