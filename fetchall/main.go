//go:build !solution

package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

func errored(e error) bool {
	isError := false
	if e != nil {
		isError = true
	}
	return isError
}

type HTTPResponseType struct {
	URL        string
	StatusCode int
	Content    string
	Elapsed    float64
	Error      error
}

func getPageContent(rawURL string) HTTPResponseType {
	t0 := time.Now()

	response := HTTPResponseType{URL: rawURL}

	url, err := url.Parse(rawURL)
	if errored(err) {
		response.Error = err
		return response
	}

	res, err := http.Get(url.String())
	if errored(err) {
		response.Error = err
		return response
	}

	body, err := io.ReadAll(res.Body)
	if errored(err) {
		response.Error = err
		return response
	}

	defer res.Body.Close()

	response.Content = string(body)
	response.StatusCode = res.StatusCode
	response.Elapsed = time.Since(t0).Seconds()

	return response
}

func ProcessHTTPRequest(url string) {
	resp := getPageContent(url)
	if resp.Error != nil {
		fmt.Printf("request to %s errored msg %s\n", resp.URL, resp.Error)
		return
	}
	fmt.Printf("%f sec\t%d\t%s\n", resp.Elapsed, resp.StatusCode, resp.URL)
}

func AsyncMap[T any](slice []T, f func(T)) {
	var wg sync.WaitGroup

	for _, obj := range slice {
		wg.Add(1)

		go func() {
			defer wg.Done()
			f(obj)
		}()
	}

	wg.Wait()
}

func FecthUrls(urls []string) {
	AsyncMap(urls, ProcessHTTPRequest)
}

func main() {
	urls := os.Args[1:]
	t0 := time.Now()

	FecthUrls(urls)
	fmt.Printf("Elapsed %f sec", time.Since(t0).Seconds())
}
