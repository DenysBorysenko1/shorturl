package pool

import "sync"

type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	pool sync.Pool
}

func New[T Resetter](factory func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				return factory()
			},
		},
	}
}

func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *Pool[T]) Put(item T) {
	item.Reset()
	p.pool.Put(item)
}
