//go:build !solution

package batcher

import (
	"sync"

	"gitlab.com/slon/shad-go/batcher/slow"
)

type Batcher struct {
	m     sync.Mutex
	v     *slow.Value
	isrun bool
	vrsn  int32
	done  chan struct{}
	res   interface{}
}

func NewBatcher(v *slow.Value) *Batcher {
	return &Batcher{v: v}
}

func (b *Batcher) SlowLoad(currentVersion int32, ready chan struct{}) {
	res := b.v.Load()

	b.m.Lock()
	b.isrun = false
	b.res = res
	close(b.done)
	b.m.Unlock()
}

func (b *Batcher) Load() interface{} {
	b.m.Lock()
	for {
		curVrsn := b.v.Version

		if b.isrun && b.vrsn >= curVrsn {
			b.m.Unlock()
			<-b.done
			b.m.Lock()
			res := b.res
			b.m.Unlock()
			return res
		}

		if !b.isrun {
			b.isrun = true
			b.vrsn = curVrsn
			b.done = make(chan struct{})
			go b.SlowLoad(curVrsn, b.done)

			b.m.Unlock()

			<-b.done

			b.m.Lock()
			res := b.res
			b.m.Unlock()
			return res
		}

		b.m.Unlock()
		<-b.done
		b.m.Lock()
	}
}
