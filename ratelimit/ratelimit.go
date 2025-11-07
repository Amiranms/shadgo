package ratelimit

import (
	"context"
	"errors"
	"time"
)

var ErrStopped = errors.New("limiter stopped")

type Mutex struct {
	locker chan struct{}
}

func NewMutex() *Mutex {
	return &Mutex{locker: make(chan struct{}, 1)}
}

func (mu *Mutex) Lock() {
	mu.locker <- struct{}{}
}

func (mu *Mutex) Unlock() {
	<-mu.locker
}

type Limiter struct {
	mu         *Mutex
	maxCount   int
	interval   time.Duration
	stop       chan struct{}
	timestamps []time.Time
}

func NewLimiter(maxCount int, interval time.Duration) *Limiter {
	if interval == 0 || maxCount == 0 {
		// без ограничений
		return &Limiter{
			stop: make(chan struct{}),
		}
	}

	return &Limiter{
		mu:         NewMutex(),
		maxCount:   maxCount,
		interval:   interval,
		stop:       make(chan struct{}),
		timestamps: make([]time.Time, 0, maxCount),
	}
}

func (l *Limiter) Acquire(ctx context.Context) error {
	if l.maxCount == 0 || l.interval == 0 {
		select {
		case <-l.stop:
			return ErrStopped
		default:
			return nil
		}
	}

	for {
		now := time.Now()
		var waitTime time.Duration

		l.mu.Lock()

		i := 0
		for i < len(l.timestamps) && now.Sub(l.timestamps[i]) >= l.interval {
			i++
		}
		l.timestamps = l.timestamps[i:]

		if len(l.timestamps) < l.maxCount {
			l.timestamps = append(l.timestamps, now)
			l.mu.Unlock()
			return nil
		}

		waitTime = l.interval - now.Sub(l.timestamps[0])
		l.mu.Unlock()

		timer := time.NewTimer(waitTime)
		select {
		case <-l.stop:
			timer.Stop()
			return ErrStopped
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (l *Limiter) Stop() {
	select {
	case <-l.stop:
	default:
		close(l.stop)
	}
}
