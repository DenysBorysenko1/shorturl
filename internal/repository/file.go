package repository

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type FileRepository[T Entity] struct {
	filename string
	mu       sync.RWMutex
}

func NewFileRepository[T Entity](filename string) (*FileRepository[T], error) {
	repo := &FileRepository[T]{
		filename: filename,
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if err := os.WriteFile(filename, []byte("[]"), 0666); err != nil {
			return nil, err
		}
	}

	return repo, nil
}

func (r *FileRepository[T]) CreateMany(entities []T) (int, error) {
	items, err := r.loadAll()
	if err != nil {
		return 0, err
	}

	items = append(items, entities...)
	err = r.saveAll(items)

	if err != nil {
		return 0, err
	}

	return len(entities), nil
}

func (r *FileRepository[T]) loadAll() ([]T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := os.ReadFile(r.filename)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return []T{}, nil
	}

	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *FileRepository[T]) saveAll(items []T) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := json.Marshal(items)
	if err != nil {
		return err
	}

	if err := os.WriteFile(r.filename, data, 0666); err != nil {
		return err
	}

	return nil
}

func (r *FileRepository[T]) Create(entity T) error {
	items, err := r.loadAll()
	if err != nil {
		return err
	}

	items = append(items, entity)

	return r.saveAll(items)
}

func (r *FileRepository[T]) GetByID(id string) (T, error) {
	items, err := r.loadAll()
	if err != nil {
		var zero T
		return zero, err
	}

	for _, item := range items {
		if item.GetID() == id {
			return item, nil
		}
	}

	var zero T
	return zero, errors.New("not found")
}
