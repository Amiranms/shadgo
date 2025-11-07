//go:build !solution

package pubsub

import (
	"context"
	"fmt"
	"sync"
)

var _ PubSub = (*MyPubSub)(nil)

var _ Subscription = (*MySubscription)(nil)

type Msg struct {
	topic string
	msg   interface{}
}

type Subscriber struct {
	c        *sync.Cond
	cb       MsgHandler
	isActive bool
	q        []interface{}
	done     chan struct{}
	owner    *MyPubSub
}

func (sub *Subscriber) deactivate() {
	sub.c.L.Lock()
	if sub.isActive {
		sub.isActive = false
		sub.c.Broadcast()
	}
	sub.c.L.Unlock()
}

func (sub *Subscriber) runcb() {
	defer close(sub.done)

	for {
		sub.c.L.Lock()
		for len(sub.q) == 0 && sub.isActive {
			sub.c.Wait()
		}

		if len(sub.q) == 0 && !sub.isActive {
			sub.c.L.Unlock()
			return
		}

		msg := sub.q[0]
		sub.q = sub.q[1:]
		sub.c.L.Unlock()

		sub.cb(msg)
		sub.owner.wg.Done()
	}
}

func (sub *Subscriber) getMsg(msg interface{}) {
	sub.c.L.Lock()
	sub.q = append(sub.q, msg)
	sub.c.Signal()
	sub.c.L.Unlock()
}

type MySubscription struct {
	topic string
	sub   *Subscriber
	pub   *MyPubSub
}

func (s *MySubscription) Unsubscribe() {
	var once sync.Once
	once.Do(func() {
		s.sub.deactivate()
		s.pub.delSub(s)
		<-s.sub.done
	})
	s.sub = nil
	s.pub = nil
}

type MyPubSub struct {
	m      sync.RWMutex
	topics map[string][]*Subscriber
	closed bool
	wg     sync.WaitGroup
}

func NewPubSub() PubSub {
	return &MyPubSub{
		topics: make(map[string][]*Subscriber),
	}
}

func (p *MyPubSub) delSub(s *MySubscription) {
	p.m.Lock()
	topicName := s.topic
	subs := p.topics[topicName]
	for i, sub := range subs {
		if sub == s.sub {
			subs[i] = subs[len(subs)-1]
			subs = subs[:len(subs)-1]
			break
		}
	}
	if len(subs) == 0 {
		delete(p.topics, topicName)
	} else {
		p.topics[topicName] = subs
	}
	p.m.Unlock()
}

func (p *MyPubSub) Subscribe(subj string, cb MsgHandler) (Subscription, error) {
	p.m.Lock()
	if p.closed {
		p.m.Unlock()
		return nil, fmt.Errorf("pubsub is closed")
	}
	sub := &Subscriber{
		c:        sync.NewCond(&sync.Mutex{}),
		cb:       cb,
		isActive: true,
		done:     make(chan struct{}),
		owner:    p,
	}
	go sub.runcb()

	p.topics[subj] = append(p.topics[subj], sub)

	subsciption := &MySubscription{
		sub:   sub,
		topic: subj,
		pub:   p,
	}

	p.m.Unlock()

	return subsciption, nil
}

func (p *MyPubSub) Publish(subj string, msg interface{}) error {
	p.m.RLock()
	if p.closed {
		p.m.RUnlock()
		return fmt.Errorf("pubsub is closed")
	}
	subs := append([]*Subscriber{}, p.topics[subj]...)
	p.m.RUnlock()

	for _, s := range subs {
		p.wg.Add(1)
		s.c.L.Lock()
		s.q = append(s.q, msg)
		s.c.Signal()
		s.c.L.Unlock()
	}
	return nil
}

func (p *MyPubSub) Close(ctx context.Context) error {
	p.m.RLock()
	var subs []*Subscriber
	for _, sl := range p.topics {
		subs = append(subs, sl...)
	}
	p.m.RUnlock()

	for _, s := range subs {
		s.deactivate()
	}

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}

	p.m.Lock()
	p.topics = nil
	p.closed = true
	p.m.Unlock()

	for _, s := range subs {
		select {
		case <-s.done:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}
