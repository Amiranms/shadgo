//go:build !solution

package tparallel

import (
	"sync"
)

type T struct {
	subtestCond    *sync.Cond
	toptestCond    *sync.Cond
	wg             *sync.WaitGroup
	topCallerBlock chan struct{}
}

func (t *T) Parallel() {
	t.executeParallel()
	t.toptestCond.L.Lock()
	defer t.toptestCond.L.Unlock()
	t.toptestCond.Wait()
}

func NewT(topCond, subCond *sync.Cond, wg *sync.WaitGroup) *T {
	return &T{
		toptestCond:    topCond,
		topCallerBlock: make(chan struct{}, 1),
		subtestCond:    subCond,
		wg:             wg,
	}
}

func (t *T) waitSequential() {
	<-t.topCallerBlock
}

func (t *T) executeParallel() {
	t.topCallerBlock <- struct{}{}
}

func (t *T) Run(subtest func(t *T)) {
	subt := NewT(t.subtestCond,
		sync.NewCond(&sync.Mutex{}),
		&sync.WaitGroup{},
	)
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		defer close(subt.topCallerBlock)
		subtest(subt)
		subt.subtestCond.Broadcast()
		subt.wg.Wait()
	}()
	subt.waitSequential()
}

func Run(topTests []func(t *T)) {
	cond := sync.NewCond(&sync.Mutex{})
	topwg := &sync.WaitGroup{}
	for _, tf := range topTests {
		NewT(nil, cond, topwg).Run(tf)
	}
	cond.Broadcast()
	topwg.Wait()
}
