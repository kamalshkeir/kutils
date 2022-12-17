package kchans

import "runtime"

func Drain[T any](c <-chan T) {
	for range c {
	}
}

func Merge[T any](c1, c2 <-chan T) <-chan T {
	r := make(chan T)
	go func(c1, c2 <-chan T, r chan<- T) {
		defer close(r)
		for c1 != nil || c2 != nil {
			select {
			case v1, ok := <-c1:
				if ok {
					r <- v1
				} else {
					c1 = nil
				}
			case v2, ok := <-c2:
				if ok {
					r <- v2
				} else {
					c2 = nil
				}
			}
		}
	}(c1, c2, r)
	return r
}

func Ranger[T any]() (*Sender[T], *Receiver[T]) {
	c := make(chan T)
	d := make(chan bool)
	s := &Sender[T]{values: c, done: d}
	r := &Receiver[T]{values: c, done: d}
	// The finalizer on the receiver will tell the sender
	// if the receiver stops listening.
	runtime.SetFinalizer(r, r.finalize)
	return s, r
}

type Sender[T any] struct {
	values chan<- T
	done   <-chan bool
}

func (s *Sender[T]) Send(v T) bool {
	select {
	case s.values <- v:
		return true
	case <-s.done:
		// The receiver has stopped listening.
		return false
	}
}

func (s *Sender[T]) Close() {
	close(s.values)
}

type Receiver[T any] struct {
	values <-chan T
	done   chan<- bool
}

func (r *Receiver[T]) Next() (T, bool) {
	v, ok := <-r.values
	return v, ok
}

func (r *Receiver[T]) finalize() {
	close(r.done)
}
