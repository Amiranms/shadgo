//go:build !solution

package lrucache

// package main

import (
	"container/list"
	"errors"
	"fmt"
)

var EmptyStack error = errors.New("Empty stack")

type Stack struct {
	l   *list.List
	cap int
}

func NewStack(cap int) *Stack {
	l := list.New()
	return &Stack{l: l, cap: cap}
}

func (s *Stack) get(k int) (*list.Element, bool) {
	if s.l.Len() == 0 {
		return nil, false
	}
	el := s.l.Front()
	for ; el != nil; el = el.Next() {
		if el.Value == k {
			return el, true
		}
	}
	return nil, false
}

func (s *Stack) PopBack() error {
	back := s.l.Back()
	if back == nil {
		return EmptyStack
	}
	s.l.Remove(back)
	return nil
}

func (s *Stack) MoveToFront(k int) {
	if el, ok := s.get(k); ok {
		s.l.MoveToFront(el)
	}
}

func (s *Stack) PushFront(value int) error {
	if s.l.Len() == s.cap {
		if err := s.PopBack(); err != nil {
			return err
		}
	}
	s.l.PushFront(value)

	return nil
}

func (s *Stack) FreeSpace() int {
	return s.cap - s.l.Len()
}

func (s *Stack) Back() (int, error) {
	if b := s.l.Back(); b != nil {
		return b.Value.(int), nil
	}
	return 0, errors.New("Empty stack")
}

func PrintList(l *list.List) {
	fmt.Println("list<", l, "> Values:")
	el := l.Front()
	for ; el != nil; el = el.Next() {
		fmt.Println(el)
	}
	fmt.Println()

}

type LRUCache struct {
	cap   int
	cache map[int]int
	stack *Stack
}

func NewLRU(cap int) LRUCache {
	cache := make(map[int]int)
	stack := NewStack(cap)
	return LRUCache{cap: cap, cache: cache, stack: stack}
}

func (c LRUCache) Range(f func(key, value int) bool) {

	for el := c.stack.l.Back(); el != nil; el = el.Prev() {
		key := el.Value.(int)
		if !f(key, c.cache[key]) {
			break
		}
	}
}

func (c *LRUCache) Clear() {
	c.cache = make(map[int]int)
	c.stack = NewStack(c.cap)
}

func (c LRUCache) Get(k int) (int, bool) {
	v, ok := c.cache[k]
	c.stack.MoveToFront(k)
	return v, ok
}

func (c LRUCache) Least() (int, error) {
	return c.stack.Back()
}

func (c LRUCache) Set(k, v int) {
	if c.cap == 0 {
		return
	}
	if _, ok := c.cache[k]; ok {
		c.stack.MoveToFront(k)
	} else {
		if c.stack.FreeSpace() == 0 {
			if least, err := c.Least(); err == nil {
				delete(c.cache, least)
			}
		}
		c.stack.PushFront(k)
	}
	c.cache[k] = v
}

func New(cap int) LRUCache {
	return NewLRU(cap)
}

// func main() {
// 	c := New(0)
// 	fmt.Println(c.Get(15))
// 	c.Set(12, 3)
// 	fmt.Println(c.Get(12))

// }
