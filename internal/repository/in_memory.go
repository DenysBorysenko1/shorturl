package repository

import (
	"errors"
	"sync"
)

type InMemoryRepository[T Entity] struct {
	data []T
	mu   sync.RWMutex
}

func NewInMemoryRepository[T Entity]() *InMemoryRepository[T] {
	return &InMemoryRepository[T]{
		data: make([]T, 0),
	}
}

func (repository *InMemoryRepository[T]) Create(entity T) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	repository.data = append(repository.data, entity)

	return nil
}

func (repository *InMemoryRepository[T]) GetById(id string) (T, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	for _, item := range repository.data {
		if item.GetID() == id {
			return item, nil
		}
	}

	var zero T

	return zero, errors.New("not found")

}
