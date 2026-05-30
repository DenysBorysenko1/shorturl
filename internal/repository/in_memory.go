package repository

import (
	"errors"
	"slices"
	"shorturl/internal/model"
	"sync"
)

type InMemoryRepository struct {
	data []model.Link
	mu   sync.RWMutex
}

func NewInMemoryRepository[T Entity]() *InMemoryRepository {
	return &InMemoryRepository{
		data: make([]model.Link, 0),
	}
}

func (repository *InMemoryRepository) CreateMany(entities []model.Link) (int, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	repository.data = append(repository.data, entities...)

	return len(entities), nil
}

func (repository *InMemoryRepository) Create(entity model.Link) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	repository.data = append(repository.data, entity)

	return nil
}

func (repository *InMemoryRepository) GetByID(id string) (model.Link, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	for _, item := range repository.data {
		if item.GetID() == id {
			return item, nil
		}
	}

	var zero model.Link

	return zero, errors.New("not found")

}

func (repository *InMemoryRepository) GetByURL(url string) (model.Link, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	for _, item := range repository.data {
		if item.GetURL() == url {
			return item, nil
		}
	}

	var zero model.Link
	return zero, errors.New("not found")
}

func (repository *InMemoryRepository) GetAllByUserID(userID string) ([]model.Link, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	result := make([]model.Link, 0, len(repository.data)/2)

	for _, item := range repository.data {
		if item.GetCreatedBy() == userID && !item.IsDeleted {
			result = append(result, item)
		}
	}

	return result, nil
}

func (repository *InMemoryRepository) SoftDeleteByIDs(ids []string, userID string) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	for i := range repository.data {
		if repository.data[i].CreatedBy == userID && slices.Contains(ids, repository.data[i].ID) {
			repository.data[i].IsDeleted = true
		}
	}

	return nil
}
