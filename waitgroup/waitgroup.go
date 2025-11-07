package waitgroup

// A WaitGroup waits for a collection of goroutines to finish.
// The main goroutine calls Add to set the number of
// goroutines to wait for. Then each of the goroutines
// runs and calls Done when finished. At the same time,
// Wait can be used to block until all goroutines have finished.
type WaitGroup chan GroupCounter

type GroupCounter struct {
	blocker chan struct{}
	count   int
}

// New creates WaitGroup.
func New() *WaitGroup {
	gctr := GroupCounter{
		blocker: make(chan struct{}),
	}
	wg := WaitGroup(make(chan GroupCounter, 1))
	wg <- gctr
	return &wg
}

func (wg *WaitGroup) lock() GroupCounter {
	return <-*wg
}

func (wg *WaitGroup) unlock(gctr GroupCounter) {
	*wg <- gctr
}

// Add adds delta, which may be negative, to the WaitGroup counter.
// If the counter becomes zero, all goroutines blocked on Wait are released.
// If the counter goes negative, Add panics.
//
// Note that calls with a positive delta that occur when the counter is zero
// must happen before a Wait. Calls with a negative delta, or calls with a
// positive delta that start when the counter is greater than zero, may happen
// at any time.
// Typically this means the calls to Add should execute before the statement
// creating the goroutine or other event to be waited for.
// If a WaitGroup is reused to wait for several independent sets of events,
// new Add calls must happen after all previous Wait calls have returned.
// See the WaitGroup example.
func (wg *WaitGroup) Add(delta int) {
	gctr := wg.lock()

	newCount := gctr.count + delta
	if newCount < 0 {
		wg.unlock(gctr)
		panic("negative WaitGroup counter")
	}

	if gctr.count == 0 && newCount > 0 {
		gctr.blocker = make(chan struct{})
	}

	gctr.count = newCount

	if gctr.count == 0 {
		close(gctr.blocker)
	}

	wg.unlock(gctr)
}

// Done decrements the WaitGroup counter by one.
func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

// Wait blocks until the WaitGroup counter is zero.
func (wg *WaitGroup) Wait() {
	gctr := wg.lock()
	blocker := gctr.blocker
	wg.unlock(gctr)

	<-blocker
}
