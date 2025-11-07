// file with ResponseExecutor
package main

import (
	"fmt"
	"net/http"
	"strings"
)

type FirewallResponseWriter struct {
	w          http.ResponseWriter
	statusCode int
	body       []byte
}

func (fw *FirewallResponseWriter) WriteHeader(code int) {
	fw.statusCode = code
}

func (fw *FirewallResponseWriter) Write(b []byte) (int, error) {
	fw.body = append(fw.body, b...)
	return len(b), nil
}

func (fw *FirewallResponseWriter) Header() http.Header {
	return fw.w.Header()
}

func (fw *FirewallResponseWriter) Send() (int, error) {
	fw.w.WriteHeader(fw.statusCode)
	return fw.w.Write(fw.body)
}

func (fw *FirewallResponseWriter) Restrict() (int, error) {
	fw.body = nil
	h := fw.w.Header()
	h.Del("Content-Length")
	h.Set("Content-Type", "text/plain")

	fw.statusCode = http.StatusForbidden
	fw.body = []byte("Forbidden")
	return fw.Send()
}

func (fw *FirewallResponseWriter) Headers() string {
	return getFullResponseHeaders(fw)
}

func (fw *FirewallResponseWriter) Body() string {
	return string(fw.body)
}

func (fw *FirewallResponseWriter) StatusCode() int {
	return fw.statusCode
}
func getFullResponseHeaders(w *FirewallResponseWriter) string {
	var headerString strings.Builder

	for name, values := range w.Header() {
		headerString.WriteString(fmt.Sprintf("%s: %s\n", name, strings.Join(values, ", ")))
	}

	return headerString.String()
}

