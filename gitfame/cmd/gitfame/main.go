//go:build !solution

package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

//go:embed language_extensions.json
var langExtsCfg []byte

type LangExts struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Extensions []string `json:"extensions"`
}

type PersonBlame struct {
	Name    string `json:"name"`
	Lines   int    `json:"lines"`
	Commits int    `json:"commits"`
	Files   int    `json:"files"`
}

type BlameStats struct {
	lines   int
	files   map[string]struct{}
	commits map[string]struct{}
}

type Revision string

func (h Revision) String() string {
	return string(h)
}

func (h Revision) Short() string {
	if len(h) >= 7 {
		return string(h)[:7]
	}
	return string(h)
}

type codeStats struct {
	lines   int
	commits int
	files   int
}

type statsCollector struct {
	sc map[string]codeStats
}

func newStatsCollector() *statsCollector {
	return &statsCollector{
		sc: make(map[string]codeStats),
	}
}

func NewPrintOredr(pOrder string) {
	//
}

type filterPatterns struct {
	exclude    []string
	include    []string
	extensions []string
	languages  []string
}

type FlamerConfig struct {
	useCommiter bool
	orderBy     string
	format      string
	filters     *filterPatterns
}

type GitFamer struct {
	RepoPath string
	Revision Revision
	config   FlamerConfig
}

func NewGitFamer(
	repoPath string,
	revision Revision,
	useCommiter bool,
	orderBy string,
	format string,
	extensions []string,
	languages []string,
	exclude []string,
	restrictTo []string,
) *GitFamer {
	filters := &filterPatterns{
		exclude:    exclude,
		include:    restrictTo,
		extensions: extensions,
		languages:  languages,
	}

	config := FlamerConfig{
		useCommiter: useCommiter,
		orderBy:     orderBy,
		format:      format,
		filters:     filters,
	}
	return &GitFamer{
		RepoPath: repoPath,
		Revision: revision,
		config:   config,
	}

}

func (gf *GitFamer) GitFiles() ([]string, error) {
	args := CreateLsFileArgs(gf.RepoPath, gf.Revision)
	return gitLsFilesSlice(args, gf.config.filters)
}

func (gf *GitFamer) GitBlameFile(file string) ([]byte, error) {
	args := CreateGitBlameArgs(gf.RepoPath, file, gf.Revision)
	cmd := exec.Command("git", args...)
	return cmd.Output()
}

func (gf *GitFamer) GitLogFile(file string) ([]byte, error) {
	args := CreateGitLogArgs(gf.RepoPath, file, gf.Revision)
	cmd := exec.Command("git", args...)
	return cmd.Output()
}

func (gf *GitFamer) CountLogStats(r io.Reader, file string) (map[string]*BlameStats, error) {
	scanner := bufio.NewScanner(r)
	authorLineIdentifier := "Author:"
	commitLineIdentifier := "commit"
	var stats map[string]*BlameStats
	var currCommit string
SCANNER:
	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.Split(line, " ")
		switch parts[0] {
		case authorLineIdentifier:
			line = strings.TrimPrefix(line, authorLineIdentifier)
			line = strings.TrimSpace(line)
			partsBeforeName := strings.SplitN(line, "<", 2)
			author := strings.TrimSpace(partsBeforeName[0])
			stats = make(map[string]*BlameStats)
			stats[author] = &BlameStats{
				files:   make(map[string]struct{}),
				commits: make(map[string]struct{}),
			}
			stats[author].files[file] = struct{}{}
			stats[author].commits[currCommit] = struct{}{}
			break SCANNER
		case commitLineIdentifier:
			currCommit = parts[1]
		default:
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}
func (gf *GitFamer) FileExists(file string) bool {
	cmd := exec.Command("git", "-C", gf.RepoPath, "cat-file", "-e", fmt.Sprintf("%s:%s", gf.Revision, file))
	return cmd.Run() == nil
}

func (gf *GitFamer) GitCountFileLines(file string) (int, error) {
	args := CreateCatFileArgs(gf.RepoPath, file, gf.Revision)
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return -1, err
	}
	trimmedStr := strings.Trim(string(out), "\n")
	count, err := strconv.Atoi(trimmedStr)
	if err != nil {
		return -1, err
	}
	return count, nil
}

func (gf *GitFamer) CountBlameStats(r io.Reader, fname string) (map[string]*BlameStats, error) {
	scanner := bufio.NewScanner(r)
	var currAuthor, currCommit string
	authorIdentifier := "author"
	if gf.config.useCommiter {
		authorIdentifier = "committer"
	}
	stats := make(map[string]*BlameStats)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "\t") {
			stats[currAuthor].lines++
			continue
		}

		parts := strings.Split(line, " ")
		switch parts[0] {
		case authorIdentifier:
			currAuthor = strings.Join(parts[1:], " ")
			var authorStats *BlameStats
			var ok bool
			if authorStats, ok = stats[currAuthor]; !ok {
				stats[currAuthor] = &BlameStats{
					files:   make(map[string]struct{}),
					commits: make(map[string]struct{}),
				}
				authorStats = stats[currAuthor]
			}
			authorStats.commits[currCommit] = struct{}{}
			authorStats.files[fname] = struct{}{}
			stats[currAuthor] = authorStats
		default:
			if len(parts) >= 3 && len(parts[0]) > 10 {
				currCommit = parts[0]
				_ = currCommit
			}
		}

	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

func (gf *GitFamer) StringBlameFile(fname string) (string, error) {
	b, err := gf.GitBlameFile(fname)
	return string(b), err
}

func (gf *GitFamer) FileStats(fname string) map[string]*BlameStats {
	lines, err := gf.GitCountFileLines(fname)
	if err != nil {
		return nil
	}
	if lines == 0 {
		return gf.LogFileStats(fname)
	}
	return gf.BlameFileStats(fname)
}

func (gf *GitFamer) BlameFileStats(fname string) map[string]*BlameStats {
	b, err := gf.GitBlameFile(fname)
	if err != nil {
		return nil
	}
	blameReader := bufio.NewReader(bytes.NewReader(b))
	fstat, err := gf.CountBlameStats(blameReader, fname)
	if err != nil {
		return nil
	}
	return fstat
}

func (bs *BlameStats) get(field string) int {
	switch field {
	case "lines":
		return bs.lines
	case "commits":
		return len(bs.commits)
	case "files":
		return len(bs.files)
	}
	return -1
}

func (gf *GitFamer) LogFileStats(fname string) map[string]*BlameStats {
	b, err := gf.GitLogFile(fname)
	if err != nil {
		return nil
	}

	logReader := bufio.NewReader(bytes.NewReader(b))
	fstat, err := gf.CountLogStats(logReader, fname)
	if err != nil {
		return nil
	}
	return fstat
}

func (gf *GitFamer) print(m map[string]*BlameStats) {
	ord := gf.config.orderBy
	switch gf.config.format {
	case "tabular":
		printTabular(prepareRecordsString(m, ord))
	case "csv":
		printCSV(prepareRecordsString(m, ord))
	case "json":
		printJSON(prepareRecordsStructs(m, ord), false)
	case "json-lines":
		printJSON(prepareRecordsStructs(m, ord), true)
	default:
		fmt.Println()
	}

}

func (gf *GitFamer) GitFame() error {
	files, err := gf.GitFiles()
	if err != nil {
		return err
	}
	stats := make(map[string]*BlameStats)
	for _, f := range files {
		mergeStats(stats, gf.FileStats(f))
	}
	gf.print(stats)
	return nil
}

func prepareRecordsString(m map[string]*BlameStats, ordedBy string) [][]string {
	records := [][]string{
		{"Name", "Lines", "Commits", "Files"},
	}
	sortedNamed := getSortedKeys(m, ordedBy)
	for _, k := range sortedNamed {
		v := m[k]
		Name := k
		Lines := strconv.Itoa(v.lines)
		Commits := strconv.Itoa(len(v.commits))
		Files := strconv.Itoa(len(v.files))
		newRec := []string{Name, Lines, Commits, Files}
		records = append(records, newRec)
	}
	return records
}

func prepareRecordsStructs(m map[string]*BlameStats, ordedBy string) []PersonBlame {
	pbs := []PersonBlame{}
	sortedNamed := getSortedKeys(m, ordedBy)
	for _, k := range sortedNamed {
		v := m[k]
		Name := k
		Lines := v.lines
		Commits := len(v.commits)
		Files := len(v.files)
		pb := PersonBlame{
			Name:    Name,
			Lines:   Lines,
			Commits: Commits,
			Files:   Files,
		}
		pbs = append(pbs, pb)
	}
	return pbs
}

func getSortedKeys(m map[string]*BlameStats, orderBy string) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	options := []string{"lines", "commit", "files"}
	for i, o := range options {
		if o == orderBy {
			options = append(options[:i], options[i:]...)
		}
	}

	sort.Slice(keys, func(i, j int) bool {
		ob := orderBy
		n1, n2 := keys[i], keys[j]
		v1, v2 := m[n1].get(ob), m[n2].get(ob)
		if v1 != v2 {
			return v1 > v2
		}
		for _, ob := range options {
			v1, v2 := m[n1].get(ob), m[n2].get(ob)
			if v1 != v2 {
				return v1 > v2
			}
		}
		return strings.ToLower(n1) < strings.ToLower(n2)
	})
	return keys
}

func printCSV(data [][]string) {
	w := csv.NewWriter(os.Stdout)
	err := w.WriteAll(data)
	if err != nil {
		panic(err)
	}
	w.Flush()
}

func printTabular(data [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	for _, row := range data {
		for i, col := range row {
			if i < len(row)-1 {
				fmt.Fprint(w, col, "\t")
			} else {
				fmt.Fprint(w, col)
			}
		}
		fmt.Fprintln(w)
	}
	w.Flush()
}

func printJSON(data []PersonBlame, lines bool) {
	if lines {
		for _, l := range data {
			ser, _ := json.Marshal(l)
			fmt.Println(string(ser))
		}
		return
	}
	ser, _ := json.Marshal(data)
	fmt.Println(string(ser))
}

// const langExtsPathDebug = "../../configs/language_extensions.json"

func getLangExtsFromEmbed() ([]LangExts, error) {
	langExtsInfo := []LangExts{}

	err := json.Unmarshal(langExtsCfg, &langExtsInfo)

	return langExtsInfo, err
}

func readLangToExtsMap() (map[string][]string, error) {
	langExtsInfo, err := getLangExtsFromEmbed()

	if err != nil {
		return nil, err
	}

	langToExts := make(map[string][]string)

	for _, l := range langExtsInfo {
		if _, ok := langToExts[l.Name]; !ok {
			langName := strings.ToLower(l.Name)
			langToExts[langName] = l.Extensions
			continue
		}
		langToExts[l.Name] = append(langToExts[l.Name], l.Extensions...)
	}
	return langToExts, nil
}

func getExts(langs []string) ([]string, error) {
	allowedExts := []string{}
	ltoe, err := readLangToExtsMap()
	if err != nil {
		return allowedExts, err
	}

	for _, l := range langs {
		lName := strings.ToLower(l)
		if e, ok := ltoe[lName]; ok {
			allowedExts = append(allowedExts, e...)
		}
	}

	return allowedExts, nil

}

func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))

	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func (fp *filterPatterns) filerFiles(fs []string) ([]string, error) {
	byLang, err := getExts(fp.languages)
	if err != nil {
		return nil, err
	}
	exts := unique(append(byLang, fp.extensions...))
	filtered := filterFilesByExts(fs, exts)

	if len(fp.include) != 0 {
		filtered = filterFilesByMatch(filtered, fp.include)
	}
	if len(fp.exclude) != 0 {
		toexclude := filterFilesByMatch(filtered, fp.exclude)
		filtered = difference(filtered, toexclude)
	}
	return filtered, nil
}

func filterFilesByMatch(files, globs []string) []string {
	matchAny := func(f string) bool {
		for _, g := range globs {
			if mtchd, _ := filepath.Match(g, f); mtchd {
				return true
			}
		}
		return false
	}

	fltrdUnique := make(map[string]struct{})
	for _, f := range files {
		if matchAny(f) {
			fltrdUnique[f] = struct{}{}
		}
	}

	fltrd := []string{}
	for f := range fltrdUnique {
		fltrd = append(fltrd, f)
	}

	return fltrd
}

func filterFilesByExts(files, exts []string) []string {
	if len(exts) == 0 {
		return files
	}

	validFiles := []string{}
	isValid := func(fname string) bool {
		for _, e := range exts {
			if filepath.Ext(fname) == e {
				return true
			}
		}
		return false
	}
	for _, f := range files {
		if isValid(f) {
			validFiles = append(validFiles, f)
		}
	}

	return validFiles
}

func unique(ss []string) []string {
	uniqer := make(map[string]struct{})

	for _, s := range ss {
		uniqer[s] = struct{}{}
	}
	uss := []string{}
	for s := range uniqer {
		uss = append(uss, s)
	}
	return uss
}

var (
	repo        string
	rev         string
	useCommiter bool
	orderBy     string
	format      string
	extensions  []string
	languages   []string
	exclude     []string
	include     []string
)

var validOrders = []string{"lines", "commits", "files"}
var validFormats = []string{"tabular", "csv", "json", "json-lines"}

var rootCmd = &cobra.Command{
	Use:   "gitfame",
	Short: "Counts repository statistics",
	Long: `Counts lines, commits, files for
	each attendant on specified git repository`,
	Run: runGitFame,
}

func init() {
	rootCmd.Flags().StringVar(&repo, "repository", ".", "path to git repository")
	rootCmd.Flags().StringVar(&rev, "revision", "HEAD", "commit revision")
	rootCmd.Flags().BoolVar(&useCommiter, "use-committer", false, "use commiter name instead of author")
	rootCmd.Flags().StringVar(&orderBy, "order-by", "lines", "stat by whic result will be ordered")
	rootCmd.Flags().StringVar(&format, "format", "tabular", "result printing format")
	rootCmd.Flags().StringSliceVar(&extensions, "extensions", []string{}, "files extensions")
	rootCmd.Flags().StringSliceVar(&languages, "languages", []string{}, "files with specified lang")
	rootCmd.Flags().StringSliceVar(&exclude, "exclude", []string{}, "files excluding patter")
	rootCmd.Flags().StringSliceVar(&include, "restrict-to", []string{}, "files including pattern")
}

func runGitFame(cmd *cobra.Command, args []string) {
	_ = args

	revision, err := GetFullRevision(repo, rev)
	if err != nil {
		os.Exit(1)
	}
	if !in(validOrders, orderBy) {
		os.Exit(1)
	}
	if !in(validFormats, format) {
		os.Exit(1)
	}

	gf := NewGitFamer(repo, revision, useCommiter,
		orderBy, format,
		extensions, languages, exclude,
		include)

	err = gf.GitFame()
	if err != nil {
		os.Exit(1)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func CreateLsFileArgs(repo string, rev Revision) []string {
	return []string{"-C",
		repo,
		"ls-tree",
		"--name-only",
		"-r",
		string(rev),
	}
}

func gitLsFilesString(args []string) (string, error) {
	cmd := exec.Command("git", args...)
	byteOutput, err := cmd.Output()
	if err != nil {
		log.Println(err)
		return "", err
	}
	return string(byteOutput), nil
}

func gitLsFilesSlice(args []string, fp *filterPatterns) ([]string, error) {
	filesStr, err := gitLsFilesString(args)
	_ = fp
	if err != nil {
		return []string{}, err
	}
	filesSlice := strings.Split(strings.TrimRight(filesStr, "\n"), "\n")
	filtered, err := fp.filerFiles(filesSlice)
	return filtered, err
}

func CreateGitBlameArgs(dir, file string, rev Revision) []string {
	return []string{"-C",
		dir,
		"blame",
		"--line-porcelain",
		string(rev),
		"--",
		file,
	}
}

func CreateCatFileArgs(dir, file string, rev Revision) []string {
	specFile := string(rev) + ":" + file
	return []string{"-C",
		dir,
		"cat-file",
		"-s",
		specFile,
	}
}

func CreateGitLogArgs(dir, file string, rev Revision) []string {
	return []string{"-C",
		dir,
		"log",
		string(rev),
		"--",
		file,
	}
}

func mergeStats(dst, src map[string]*BlameStats) {
	if dst == nil {
		return
	}
	for name, st := range src {
		dstSt, ok := dst[name]
		if !ok {
			dst[name] = st
			continue
		}
		maps.Copy(dstSt.files, st.files)
		maps.Copy(dstSt.commits, st.commits)
		dstSt.lines += st.lines
	}
}

func in(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func GetFullRevision(repo, revision string) (Revision, error) {
	args := []string{"-C", repo, "rev-parse", "--verify", revision}
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return Revision(""), fmt.Errorf("invalid revision: %s", revision)
	}
	hash := strings.TrimSpace(string(output))
	if len(hash) != 40 || !isValidSHA1(hash) {
		return Revision(""), fmt.Errorf("invalid SHA-1 hash: %s", hash)
	}
	return Revision(hash), nil
}

func isValidSHA1(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
