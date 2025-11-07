//go:build !solution

package ciletters

import (
	"bufio"
	"bytes"
	"container/list"
	_ "embed"
	"strings"
	"text/template"
)

//go:embed template.txt
var ciTemplate string

func comitHashShortener(s string, n int) string {
	return s[:n]
}

func trimRunnerLog(s string) string {
	str, _ := strings.CutPrefix(s, "$ ")
	return str
}

type StringCache struct {
	list    *list.List
	maxSize int
}

func NewStringCache(maxSize int) *StringCache {
	return &StringCache{
		list:    list.New(),
		maxSize: maxSize,
	}
}

func (c *StringCache) Add(s string) {
	if c.list.Len() >= c.maxSize {
		c.list.Remove(c.list.Front())
	}
	c.list.PushBack(s)
}

func (c *StringCache) GetAll() []string {
	var result []string
	for e := c.list.Front(); e != nil; e = e.Next() {
		result = append(result, e.Value.(string))
	}
	return result
}

func (c *StringCache) isempty() bool {
	return c.list.Len() == 0
}

func (c *StringCache) clear() {
	c.list = list.New()
}

func addLine(sb *strings.Builder, line string, prefix, lineSeparator string) {
	sb.WriteString(prefix + line + lineSeparator)
}

func addLines(sb *strings.Builder, lines []string, prefix, lineSeparator string) {
	for _, l := range lines {
		addLine(sb, l, prefix, lineSeparator)
	}
}

func processLineByLine(scanner *bufio.Scanner, prefix, lineSeparator string) (string, error) {
	var sb strings.Builder
	testToolCache := NewStringCache(3)
	onTesttool := false
	for scanner.Scan() {
		line := scanner.Text() // Get the text of the current line
		if strings.Contains(line, "testtool") {
			testToolCache.Add(line)
			onTesttool = true
			continue
		}
		if onTesttool {
			toolLogs := testToolCache.GetAll()
			addLines(&sb, toolLogs, prefix, lineSeparator)
			testToolCache.clear()
			onTesttool = false
		}
		addLine(&sb, line, prefix, lineSeparator)
	}
	return sb.String(), nil
}

func prepocessLog(str string, shiftSpaces int) string {
	morpheme := " "
	lineSeparator := "\n"
	reader := strings.NewReader(str)
	scanner := bufio.NewScanner(reader)
	prefix := strings.Repeat(morpheme, shiftSpaces)

	resultLog, err := processLineByLine(scanner, prefix, lineSeparator)

	if err != nil {
		panic(err)
	}
	return resultLog
}

func MakeLetter(notif *Notification) (string, error) {
	tmpl := template.New("main-ci")
	tmpl.Funcs(template.FuncMap{
		"shortHash":    comitHashShortener,
		"trimLog":      trimRunnerLog,
		"prepocessLog": prepocessLog,
	},
	)
	tmpl.Option("missingkey=error")
	tmpl, err := tmpl.Parse(ciTemplate)

	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, notif)
	if err != nil {
		panic(err)
	}

	return buf.String(), nil
}
