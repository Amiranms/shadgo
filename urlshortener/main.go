//go:build !solution

package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	_ "fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const defaultPort = 8889

var ShortedUrls map[string]string

func contains(slice []string, str string) (int, bool) {
	for i, s := range slice {
		if s == str {
			return i, true
		}
	}
	return -1, false
}

func getPortFromCLArgs() int {
	port := defaultPort
	args := os.Args
	if idx, ok := contains(args, "-port"); ok {
		var err error
		port, err = strconv.Atoi(args[idx+1])
		if err != nil {
			port = defaultPort
		}
	}
	return port
}

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello word")
}

func generateShortHash(url string, length int) string {
	h := sha1.New()
	h.Write([]byte(url))
	hash := hex.EncodeToString(h.Sum(nil))
	return hash[:length]
}

func getShortHashFromURL(u string) string {
	h := generateShortHash(u, 6)
	ShortedUrls[h] = u
	return h
}

type ShortenBody struct {
	Url string `json:"url"`
}

type ShortenResponse struct {
	Url string `json:"url"`
	Key string `json:"key"`
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var shrtB ShortenBody
	err := json.NewDecoder(r.Body).Decode(&shrtB)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	url := shrtB.Url
	key := getShortHashFromURL(url)
	resp := ShortenResponse{url, key}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}

func GoHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	parts := strings.Split(url, "/")
	if len(parts) != 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	key := parts[2]

	if _, ok := ShortedUrls[key]; !ok {
		http.Error(w, "Invalid path", http.StatusNotFound)
		return
	}
	redirectTo := ShortedUrls[key]
	http.Redirect(w, r, redirectTo, http.StatusFound)
}

func main() {
	ShortedUrls = make(map[string]string)
	portInt := getPortFromCLArgs()
	portStr := fmt.Sprintf(":%d", portInt)
	http.HandleFunc("/", helloWorldHandler)
	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/go/", GoHandler)

	http.ListenAndServe(portStr, nil)
}
