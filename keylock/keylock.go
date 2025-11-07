//go:build !solution

package keylock

import (
	"sort"
	"sync"
)

type KeyLock struct {
	mu   sync.Mutex
	q    []*LockRequest
	hold map[string]struct{}
}

type LockRequest struct {
	keys   []string
	notify chan struct{}
	once   sync.Once
}

func New() *KeyLock {
	return &KeyLock{
		hold: make(map[string]struct{}),
		q:    nil,
	}
}

func uniqueSortedKeys(ks []string) []string {
	set := make(map[string]struct{}, len(ks))
	for _, k := range ks {
		set[k] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}

	sort.Strings(out)
	return out
}

func (l *KeyLock) canLock(r *LockRequest) bool {
	for _, k := range r.keys {
		if _, ok := l.hold[k]; ok {
			return false
		}
	}

	return true
}

func (l *KeyLock) lock(r *LockRequest) {
	for _, k := range r.keys {
		l.hold[k] = struct{}{}
	}

	r.once.Do(func() {
		close(r.notify)
	})
}

func (l *KeyLock) serveQueue() {
	if len(l.q) == 0 {
		return
	}

	newq := l.q[:0]
	for _, r := range l.q {
		if l.canLock(r) {
			l.lock(r)
			continue
		}
		newq = append(newq, r)
	}
	l.q = newq
}

func (l *KeyLock) unlock(ks []string) {
	for _, k := range ks {
		delete(l.hold, k)
	}
}

func (l *KeyLock) popReq(r *LockRequest) {
	newq := l.q[:0]
	for _, lr := range l.q {
		if lr == r {
			continue
		}
		newq = append(newq, lr)
	}
	l.q = newq
}

func (l *KeyLock) LockKeys(keys []string, cancel <-chan struct{}) (canceled bool, unlock func()) {
	uniqueSortedKeys := uniqueSortedKeys(keys)
	r := &LockRequest{
		keys:   uniqueSortedKeys,
		notify: make(chan struct{}),
	}
	l.mu.Lock()
	if l.canLock(r) {
		l.lock(r)
		l.mu.Unlock()

		var once sync.Once
		unlock := func() {
			once.Do(func() {
				l.mu.Lock()
				l.unlock(r.keys)
				l.serveQueue()
				l.mu.Unlock()
			})
		}
		return false, unlock
	}

	l.q = append(l.q, r)
	l.mu.Unlock()

	select {
	case <-cancel:
		l.mu.Lock()
		l.popReq(r)
		l.mu.Unlock()
		return true, nil

	case <-r.notify:
		var once sync.Once
		unlock := func() {
			once.Do(func() {
				l.mu.Lock()
				l.unlock(r.keys)
				l.serveQueue()
				l.mu.Unlock()
			})
		}
		return false, unlock

	}

}
