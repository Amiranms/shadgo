// File to implement RulesExecutor interface defined in firewall main component
package main

// CCP principle implemented

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type ResponseExecutor interface {
	WriteHeader(code int)
	Write(b []byte) (int, error)
	Header() http.Header
	Headers() string
	StatusCode() int
	Body() string
	Send() (int, error)
	Restrict() (int, error)
}

type YAMLReader interface {
	Readall() ([]byte, error)
}

type Rule struct {
	Endpoint               string   `yaml:"endpoint"`
	ForbiddenUserAgents    []string `yaml:"forbidden_user_agents"`
	ForbiddenHeaders       []string `yaml:"forbidden_headers"`
	RequiredHeaders        []string `yaml:"required_headers"`
	MaxRequestLengthBytes  int      `yaml:"max_request_length_bytes"`
	MaxResponseLengthBytes int      `yaml:"max_response_length_bytes"`
	ForbiddenResponseCodes []int    `yaml:"forbidden_response_codes"`
	ForbiddenRequestRe     []string `yaml:"forbidden_request_re"`
	ForbiddenResponseRe    []string `yaml:"forbidden_response_re"`

	forbiddenUserAgentsCompiled []*regexp.Regexp
	forbiddenHeadersCompiled    []*regexp.Regexp
	forbiddenRequestReCompiled  []*regexp.Regexp
	forbiddenResponseReCompiled []*regexp.Regexp
}

func (rl *Rule) checkUserAgent(r *http.Request) bool {
	ua := r.UserAgent()
	for _, re := range rl.forbiddenUserAgentsCompiled {
		if re.MatchString(ua) {
			return false
		}
	}
	return true
}

func (rl *Rule) checkRequestForbiddenHeaders(r *http.Request) bool {
	headers := getFullRequestHeaders(r)
	for _, re := range rl.forbiddenHeadersCompiled {
		if re.MatchString(headers) {
			return false
		}
	}
	return true
}

func (rl *Rule) checkResponseForbiddenHeaders(w ResponseExecutor) bool {
	headers := w.Headers()
	for _, re := range rl.forbiddenHeadersCompiled {
		if re.MatchString(headers) {
			return false
		}
	}
	return true
}

func (rl *Rule) checkRequestRequiredHeaders(r *http.Request) bool {
	return rl.checkRequiredHeaders(r.Header)
}

func (rl *Rule) checkResponseRequiredHeaders(w ResponseExecutor) bool {
	return rl.checkRequiredHeaders(w.Header())
}

func (rl *Rule) checkRequiredHeaders(headers http.Header) bool {
	for _, h := range rl.RequiredHeaders {
		if headers.Get(h) == "" {
			return false
		}
	}
	return true
}

func (rl *Rule) checkReqContentLength(r *http.Request) bool {

	if rl.MaxRequestLengthBytes == 0 {
		return true
	}

	return rl.MaxRequestLengthBytes >= int(r.ContentLength)
}

func (rl *Rule) checkRespContentLength(w ResponseExecutor) bool {

	if rl.MaxResponseLengthBytes == 0 {
		return true
	}
	A := w.Header().Get("Content-Length")
	i, err := strconv.Atoi(A)
	if err != nil {
		fmt.Println("failed Response Content-Length check")
		return false
	}
	return rl.MaxResponseLengthBytes >= i
}

func (rl *Rule) checkReqBodyContent(r *http.Request) bool {
	pass := true
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return pass
	}

	bodyStr := string(body)
	r.Body.Close()
	for _, re := range rl.forbiddenRequestReCompiled {
		if re.MatchString(bodyStr) {
			pass = false
			break
		}
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	return pass
}

func (rl *Rule) checkRespBodyContent(w ResponseExecutor) bool {
	bodyStr := w.Body()
	for _, re := range rl.forbiddenResponseReCompiled {
		if re.MatchString(bodyStr) {
			return false
		}
	}
	return true
}

func (rl *Rule) checkResponseForbiddenCodes(w ResponseExecutor) bool {
	for _, code := range rl.ForbiddenResponseCodes {
		if code == w.StatusCode() {
			return false
		}
	}
	return true
}

func (rl *Rule) CheckRequest(r *http.Request) bool {
	if rl == nil {
		return true
	}

	if ok := rl.checkUserAgent(r); !ok {
		return false
	}

	if ok := rl.checkRequestForbiddenHeaders(r); !ok {
		return false
	}

	if ok := rl.checkRequestRequiredHeaders(r); !ok {
		return false
	}

	if ok := rl.checkReqContentLength(r); !ok {
		return false
	}

	if ok := rl.checkReqBodyContent(r); !ok {
		return false
	}

	return true
}

func (rl *Rule) CheckResponse(w ResponseExecutor) bool {
	if rl == nil {
		return true
	}

	if ok := rl.checkResponseForbiddenHeaders(w); !ok {
		return false
	}

	if ok := rl.checkResponseRequiredHeaders(w); !ok {
		return false
	}

	if ok := rl.checkRespContentLength(w); !ok {
		return false
	}

	if ok := rl.checkRespBodyContent(w); !ok {
		return false
	}

	if ok := rl.checkResponseForbiddenCodes(w); !ok {
		return false
	}

	return true
}

type RulesExecutorYaml struct {
	Rules       []*Rule `yaml:"rules"`
	r           YAMLReader
	currentRule *Rule
	w           ResponseExecutor
}

func NewRulesYaml(r YAMLReader) *RulesExecutorYaml {
	return &RulesExecutorYaml{r: r}
}

func (ru *RulesExecutorYaml) CompileRules() {
	for _, r := range ru.Rules {
		for _, fua := range r.ForbiddenUserAgents {
			r.forbiddenUserAgentsCompiled = append(
				r.forbiddenUserAgentsCompiled,
				regexp.MustCompile(fua),
			)

		}

		for _, fh := range r.ForbiddenHeaders {
			r.forbiddenHeadersCompiled = append(
				r.forbiddenHeadersCompiled,
				regexp.MustCompile(fh),
			)
		}

		for _, freq := range r.ForbiddenRequestRe {
			r.forbiddenRequestReCompiled = append(
				r.forbiddenRequestReCompiled,
				regexp.MustCompile(freq),
			)
		}

		for _, fres := range r.ForbiddenResponseRe {
			r.forbiddenResponseReCompiled = append(
				r.forbiddenResponseReCompiled,
				regexp.MustCompile(fres),
			)
		}

	}
}

func (ru *RulesExecutorYaml) ParseRules() error {
	content, err := ru.r.Readall()
	if err != nil {
		return fmt.Errorf("read from YAMLReader errored: %w", err)
	}
	err = yaml.Unmarshal(content, &ru)
	if err != nil {
		return fmt.Errorf("rules parsing errored: %w", err)
	}

	return nil
}

//	func (ru *RulesExecutorYaml) checkUserAgent(string) bool{
//		ru.forbiddenUserAgentsCompiled
//	}

func (ru *RulesExecutorYaml) getRule(r *http.Request) *Rule {
	return findBestMatch(r, ru.Rules)
}

func (ru *RulesExecutorYaml) CheckRequest(r *http.Request) bool {
	rule := ru.currentRule
	success := rule.CheckRequest(r)
	// fmt.Printf("REQUEST RULE CHECK: %t\n", success)

	return success
}

func (ru *RulesExecutorYaml) SetRule(r *http.Request) error {
	rule := ru.getRule(r)
	if rule == nil {
		return fmt.Errorf("Rule not found")
	}
	ru.currentRule = rule
	return nil
}

func (ru *RulesExecutorYaml) CheckResponse(w http.ResponseWriter) bool {
	frw, ok := w.(ResponseExecutor)
	if !ok {
		panic("Create ResponseWriterWith NewResponseWriter interface method")
	}
	rule := ru.currentRule
	success := rule.CheckResponse(frw)
	// fmt.Printf("RESPONSE RULE CHECK: %t\n", success)

	return success
}

func (ru *RulesExecutorYaml) Restrict() (int, error) {
	return ru.w.Restrict()
}

func (ru *RulesExecutorYaml) Send() (int, error) {
	return ru.w.Send()
}

// зависимость, которой хотелось бы избежать. Т.е. сейчас RulesExecutorYaml знает что такое FirewallResponseWriter, хотя ему не следовало бы
func newResponseExecutor(w http.ResponseWriter) ResponseExecutor {
	return &FirewallResponseWriter{w: w}
}

func (ru *RulesExecutorYaml) NewResponseWriter(w http.ResponseWriter) http.ResponseWriter {
	fw := newResponseExecutor(w)
	ru.w = fw
	return fw
}

// internals

func findBestMatch(r *http.Request, rules []*Rule) *Rule {
	currentPath := r.URL.Path
	var bestMatch *Rule
	var bestMatchLength int = -1

	for _, rule := range rules {

		if strings.HasPrefix(currentPath, rule.Endpoint) {
			if len(rule.Endpoint) > bestMatchLength {
				bestMatchLength = len(rule.Endpoint)
				bestMatch = rule
			}
		}
	}

	return bestMatch
}

func getFullRequestHeaders(r *http.Request) string {
	var headerString strings.Builder

	for name, values := range r.Header {
		headerString.WriteString(fmt.Sprintf("%s: %s\n", name, strings.Join(values, ", ")))
	}

	return headerString.String()
}
