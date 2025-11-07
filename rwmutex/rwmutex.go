//go:build !solution

package rwmutex

// package main

// A RWMutex is a reader/writer mutual exclusion lock.
// The lock can be held by an arbitrary number of readers or a single writer.
// The zero value for a RWMutex is an unlocked mutex.
//
// If a goroutine holds a RWMutex for reading and another goroutine might
// call Lock, no goroutine should expect to be able to acquire a read lock
// until the initial read lock is released. In particular, this prohibits
// recursive read locking. This is to ensure that the lock eventually becomes
// available; a blocked Lock call excludes new readers from acquiring the
// lock.
type RWMutex struct {
	readers int
	rLock   chan struct{}
	wLock   chan struct{}
}

// New creates *RWMutex.
func New() *RWMutex {
	rwmt := &RWMutex{
		rLock: make(chan struct{}, 1),
		wLock: make(chan struct{}, 1),
	}
	return rwmt
}

// RLock locks rw for reading.
//
// It should not be used for recursive read locking; a blocked Lock
// call excludes new readers from acquiring the lock. See the
// documentation on the RWMutex type.
func (rw *RWMutex) RLock() {
	rw.rLock <- struct{}{}
	rw.readers++
	if rw.readers == 1 {
		rw.wLock <- struct{}{}
	}
	<-rw.rLock
}

// RUnlock undoes a single RLock call;
// it does not affect other simultaneous readers.
// It is a run-time error if rw is not locked for reading
// on entry to RUnlock.
func (rw *RWMutex) RUnlock() {
	rw.rLock <- struct{}{}
	if rw.readers == 0 {
		<-rw.rLock
		panic("RWMutex is not locked")
	}
	if rw.readers == 1 {
		<-rw.wLock
	}
	rw.readers--
	<-rw.rLock
}

// Lock locks rw for writing.
// If the lock is already locked for reading or writing,
// Lock blocks until the lock is available.
func (rw *RWMutex) Lock() {
	rw.wLock <- struct{}{}
}

// Unlock unlocks rw for writing. It is a run-time error if rw is
// not locked for writing on entry to Unlock.
//
// As with Mutexes, a locked RWMutex is not associated with a particular
// goroutine. One goroutine may RLock (Lock) a RWMutex and then
// arrange for another goroutine to RUnlock (Unlock) it.
func (rw *RWMutex) Unlock() {
	<-rw.wLock
}
