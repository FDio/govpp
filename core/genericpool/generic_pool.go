package genericpool

import (
	"sync"
)

type Pool[T any] struct {
	p sync.Pool
}

func (p *Pool[T]) Get() T {
	return p.p.Get().(T)
}

func (p *Pool[T]) Put(x T) {
	p.p.Put(x)
}

func New[T any](f func() T) *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{
			New: func() any { return f() },
		},
	}
}
