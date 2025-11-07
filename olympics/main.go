//go:build !solution

package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

//NB: это говно, надо как-то обеспечить неизменяемость переменной olympic

var olympics []Olympic

func readData(path string) ([]Olympic, error) {
	rawdata, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var o []Olympic
	err = json.Unmarshal(rawdata, &o)
	if err != nil {
		return nil, err
	}
	return o, nil
}

const (
	defaultPort    = "8889"
	defaulDataPath = "testdata/olympicWinners.json"
)

type Olympic struct {
	Age     int    `json:"age"`
	Athlete string `json:"athlete"`
	Date    string `json:"date"`
	Bronze  int    `json:"bronze"`
	Silver  int    `json:"silver"`
	Gold    int    `json:"gold"`
	Year    int    `json:"year"`
	Total   int    `json:"total"`
	Sport   string `json:"sport"`
	Country string `json:"country"`
}

// {"athlete":"Michael Phelps","age":23,"country":"United States","year":2008,"date":"24/08/2008","sport":"Swimming","gold":8,"silver":0,"bronze":0,"total":8}

type Medals struct {
	Gold   int `json:"gold"`
	Silver int `json:"silver"`
	Bronze int `json:"bronze"`
	Total  int `json:"total"`
}

type Year string

func NewYear(i int) Year {
	return Year(strconv.Itoa(i))
}

type CountryInfo struct {
	Country string `json:"country"`
	Gold    int    `json:"gold"`
	Silver  int    `json:"silver"`
	Bronze  int    `json:"bronze"`
	Total   int    `json:"total"`
}

func newCountryInfo(country string) *CountryInfo {
	return &CountryInfo{
		Country: country,
	}
}

func (c *CountryInfo) countMedals(gold, silver, bronze int) {
	c.Total += gold + silver + bronze
	c.Gold += gold
	c.Silver += silver
	c.Bronze += bronze
}

func (c *CountryInfo) updateMedals(Records []Olympic) {
	for _, r := range Records {
		c.addRecord(r)
	}
}

func (c *CountryInfo) addRecord(r Olympic) {
	c.countMedals(r.Gold, r.Silver, r.Bronze)
}

type AthleteInfo struct {
	Athlete      string          `json:"athlete"`
	Country      string          `json:"country"`
	Medals       Medals          `json:"medals"`
	MedalsByYear map[Year]Medals `json:"medals_by_year"`
}

func newAthleteInfo(name, country string) *AthleteInfo {
	return &AthleteInfo{
		Athlete:      name,
		Country:      country,
		MedalsByYear: make(map[Year]Medals),
	}
}

func getAthleteInfo(olympics []Olympic, name string) (*AthleteInfo, error) {
	athletRecords := filterByAthlete(olympics, name)
	if len(athletRecords) == 0 {
		return nil, errors.New("athlete " + name + " not found")
	}
	a := newAthleteInfo(name, athletRecords[0].Country)
	a.updateMedals(athletRecords)
	return a, nil
}

func (a *AthleteInfo) countMedals(gold, silver, bronze, year int) {
	total := gold + silver + bronze
	medal := Medals{gold, silver, bronze, total}
	y := NewYear(year)
	a.Medals = AddMedals(medal, a.Medals)
	a.MedalsByYear[y] = AddMedals(medal, a.MedalsByYear[y])
}

func (a *AthleteInfo) updateMedals(athleteRecords []Olympic) {
	for _, r := range athleteRecords {
		a.addRecord(r)
	}
}

func (a *AthleteInfo) addRecord(r Olympic) {
	a.countMedals(r.Gold, r.Silver, r.Bronze, r.Year)
}

func filterByAthlete(olympics []Olympic, name string) []Olympic {
	var filtered []Olympic

	for _, o := range olympics {
		if o.Athlete == name {
			filtered = append(filtered, o)
		}
	}

	return filtered
}

func AddMedals(a, b Medals) (sum Medals) {
	sum = Medals{
		Gold:   a.Gold + b.Gold,
		Silver: a.Silver + b.Silver,
		Bronze: a.Bronze + b.Bronze,
		Total:  a.Total + b.Total,
	}
	return sum
}

func getAthleteInfoHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name not specified", http.StatusBadRequest)
		return
	}

	ai, err := getAthleteInfo(olympics, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	response, err := json.Marshal(*ai)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(response))
}

func AreEqual(s1, s2 string) bool {
	return strings.ToLower(s1) == strings.ToLower(s2)
}

func filterBySport(olympics []Olympic, sport string) []Olympic {
	var filtered []Olympic

	for _, o := range olympics {
		if AreEqual(o.Sport, sport) {
			filtered = append(filtered, o)
		}
	}

	return filtered
}

func getNames(olympics []Olympic) []string {
	set := map[string]struct{}{}
	for _, o := range olympics {
		set[o.Athlete] = struct{}{}
	}

	var unique []string
	for k := range set {
		unique = append(unique, k)
	}
	return unique
}

func joinByNames(olympics []Olympic) map[string]*AthleteInfo {
	table := map[string]*AthleteInfo{}
	for _, o := range olympics {
		name := o.Athlete
		if _, ok := table[name]; ok {
			info := table[name]
			info.addRecord(o)
			continue
		}
		table[name] = newAthleteInfo(name, o.Country)
		table[name].addRecord(o)
	}
	return table
}

func AthleteMapToList(m map[string]*AthleteInfo) []AthleteInfo {
	var list []AthleteInfo
	for _, v := range m {
		list = append(list, *v)
	}
	return list
}

func getAllInSport(olympics []Olympic, sport string) []AthleteInfo {
	fltr := filterBySport(olympics, sport)
	allMap := joinByNames(fltr)
	return AthleteMapToList(allMap)
}

func compareAthletes(a1, a2 AthleteInfo) bool {
	if a1.Medals.Gold != a2.Medals.Gold {
		return a1.Medals.Gold > a2.Medals.Gold
	}
	if a1.Medals.Silver != a2.Medals.Silver {
		return a1.Medals.Silver > a2.Medals.Silver
	}
	if a1.Medals.Bronze != a2.Medals.Bronze {
		return a1.Medals.Bronze > a2.Medals.Bronze
	}
	return a1.Athlete < a2.Athlete
}

func sortAthleteList(l []AthleteInfo) {
	sort.Slice(l, func(i, j int) bool {
		return compareAthletes(l[i], l[j])
	})
}

func getTopNInSport(olympics []Olympic, sport string, limitN int) ([]AthleteInfo, error) {
	all := getAllInSport(olympics, sport)
	if len(all) == 0 {
		return []AthleteInfo{}, errors.New("sport '" + sport + "' not found")
	}
	sortAthleteList(all)
	return all[:limitN], nil
}

func getTopNInSportXHandler(w http.ResponseWriter, r *http.Request) {
	sport := r.URL.Query().Get("sport")
	if sport == "" {
		http.Error(w, "sport have to be specified", http.StatusBadRequest)
		return
	}
	limit := r.URL.Query().Get("limit")
	var limitN = 3
	var err error
	if limit != "" {
		limitN, err = strconv.Atoi(limit)
	}
	if err != nil {
		http.Error(w, "Invalid limit", http.StatusBadRequest)
	}

	result, err := getTopNInSport(olympics, sport, limitN)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response, err := json.Marshal(result)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(response))
}

func compareCountries(c1, c2 CountryInfo) bool {
	if c1.Gold != c2.Gold {
		return c1.Gold > c2.Gold
	}
	if c1.Silver != c2.Silver {
		return c1.Silver > c2.Silver
	}
	if c1.Bronze != c2.Bronze {
		return c1.Bronze > c2.Bronze
	}
	return c1.Country < c2.Country
}

func sortCountryList(l []CountryInfo) {
	sort.Slice(l, func(i, j int) bool {
		return compareCountries(l[i], l[j])
	})
}

func joinByCountry(olympics []Olympic) map[string]*CountryInfo {
	table := map[string]*CountryInfo{}
	for _, o := range olympics {
		country := o.Country
		if _, ok := table[country]; ok {
			info := table[country]
			info.addRecord(o)
			continue
		}
		table[country] = newCountryInfo(country)
		table[country].addRecord(o)
	}
	return table
}

func filterByYear(olympics []Olympic, year int) []Olympic {
	var filtered []Olympic

	for _, o := range olympics {
		if o.Year == year {
			filtered = append(filtered, o)
		}
	}

	return filtered
}

func CountryMapToList(m map[string]*CountryInfo) []CountryInfo {
	var list []CountryInfo
	for _, v := range m {
		list = append(list, *v)
	}
	return list
}

func getAllInYear(olympics []Olympic, year int) []CountryInfo {
	f := filterByYear(olympics, year)
	allMap := joinByCountry(f)
	return CountryMapToList(allMap)
}

func getTopNCountriesInYear(olympics []Olympic, year, limit int) ([]CountryInfo, error) {
	all := getAllInYear(olympics, year)
	if len(all) == 0 {
		errStr := fmt.Sprintf("year %d not found", year)
		return []CountryInfo{}, errors.New(errStr)
	}
	sortCountryList(all)
	if limit > len(all) {
		limit = len(all)
	}
	return all[:limit], nil
}

func getTopNCountriesInYearHandler(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		http.Error(w, "year has to be specified", http.StatusBadRequest)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Invalid year", http.StatusBadRequest)
		return
	}

	limit := r.URL.Query().Get("limit")
	var limitN = 3
	if limit != "" {
		limitN, err = strconv.Atoi(limit)
	}
	if err != nil {
		http.Error(w, "Invalid limit", http.StatusBadRequest)
		return
	}

	result, err := getTopNCountriesInYear(olympics, year, limitN)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response, err := json.Marshal(result)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, string(response))
}

func main() {

	port := flag.String("port", defaultPort, "port to listen and serve")
	data := flag.String("data", defaulDataPath, "data path")
	flag.Parse()

	var err error
	olympics, err = readData(*data)
	if err != nil {
		fmt.Println(err)
		return
	}
	// fmt.Println(getTopNCountriesInYear(olympics, 2000, 104))

	m := chi.NewRouter()
	m.Get("/athlete-info", getAthleteInfoHandler)
	m.Get("/top-athletes-in-sport", getTopNInSportXHandler)
	m.Get("/top-countries-in-year", getTopNCountriesInYearHandler)
	log.Fatal(http.ListenAndServe(":"+*port, m))
}
