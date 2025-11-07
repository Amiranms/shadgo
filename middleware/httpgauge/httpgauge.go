//go:build !solution

package httpgauge

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
)

type Gauge struct {
	m     sync.Mutex
	count map[string]int
}

func New() *Gauge {
	return &Gauge{
		count: make(map[string]int),
	}
}

func (g *Gauge) Snapshot() map[string]int {
	snapshot := make(map[string]int)
	g.m.Lock()
	for k, v := range g.count {
		snapshot[k] = v
	}
	g.m.Unlock()
	return snapshot
}

// ServeHTTP returns accumulated statistics in text format ordered by pattern.
//
// For example:
//
//	/a 10
//	/b 5
//	/c/{id} 7
func (g *Gauge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	count := g.Snapshot()
	var ptrns []string

	for k := range count {
		ptrns = append(ptrns, k)
	}
	var b strings.Builder
	slices.Sort(ptrns)

	for _, p := range ptrns {
		s := p + " " + strconv.Itoa(count[p]) + "\n"
		b.WriteString(s)
	}
	fmt.Fprint(w, b.String())
}

func (g *Gauge) Count(ptrn string) {
	g.m.Lock()
	g.count[ptrn]++
	g.m.Unlock()
}

func (g *Gauge) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ptrn := chi.RouteContext(r.Context()).RoutePattern()
			g.Count(ptrn)
		}()
		next.ServeHTTP(w, r)
	})
}
