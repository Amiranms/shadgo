// !solution

package dupcall

import (
	"context"
	"errors"
	"sync"
)

type Call struct {
	dups    int
	mu      sync.Mutex
	waiters []chan Result
	fctx    context.Context
	cnclf   context.CancelFunc
}

type Result struct {
	r   interface{}
	err error
}

func removeChannelByIndex(channels []chan Result, index int) []chan Result {
	return append(channels[:index], channels[index+1:]...)
}

func removeChannelByValue(channels []chan Result, targetChan chan Result) []chan Result {
	for i, ch := range channels {
		if ch == targetChan {
			return removeChannelByIndex(channels, i)
		}
	}
	return channels
}

func (c *Call) removeFromWaiters(targetChan chan Result) {
	c.waiters = removeChannelByValue(c.waiters, targetChan)
	close(targetChan)
}

func (c *Call) Do(ctx context.Context, cb func(context.Context) (interface{}, error)) (result interface{}, err error) {
	c.mu.Lock()
	waiterChan := make(chan Result)
	c.waiters = append(c.waiters, waiterChan)
	if c.dups == 0 {
		c.fctx, c.cnclf = context.WithCancel(context.Background())
		go func() {
			res, err := cb(c.fctx)

			for _, w := range c.waiters {
				w <- Result{res, err}
			}
		}()
	}
	c.dups++
	c.mu.Unlock()
	select {
	case <-ctx.Done():
		c.mu.Lock()
		c.removeFromWaiters(waiterChan)
		if c.dups == 1 {
			c.cnclf()
		}
		c.dups--
		c.mu.Unlock()
		return nil, errors.New("canceled")
	case res := <-waiterChan:
		c.mu.Lock()
		c.removeFromWaiters(waiterChan)
		c.dups--
		c.mu.Unlock()
		return res.r, res.err
	}
}
