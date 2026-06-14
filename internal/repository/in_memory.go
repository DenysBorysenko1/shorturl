package repository

import (
	"errors"
	"shorturl/internal/model"
	"sync"
)

type InMemoryRepository struct {
	data      []model.Link
	index     map[string]model.Link
	userIndex map[string][]int
	mu        sync.RWMutex
}

func NewInMemoryRepository[T Entity]() *InMemoryRepository {
	return &InMemoryRepository{
		data:      make([]model.Link, 0),
		index:     make(map[string]model.Link),
		userIndex: make(map[string][]int),
	}
}

func (repository *InMemoryRepository) CreateMany(entities []model.Link) (int, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	startIdx := len(repository.data)
	repository.data = append(repository.data, entities...)

	for i, entity := range entities {
		repository.index[entity.GetID()] = entity
		userID := entity.GetCreatedBy()
		repository.userIndex[userID] = append(repository.userIndex[userID], startIdx+i)
	}

	return len(entities), nil
}

func (repository *InMemoryRepository) Create(entity model.Link) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	idx := len(repository.data)
	repository.data = append(repository.data, entity)
	repository.index[entity.GetID()] = entity
	repository.userIndex[entity.GetCreatedBy()] = append(repository.userIndex[entity.GetCreatedBy()], idx)

	return nil
}

func (repository *InMemoryRepository) GetByID(id string) (model.Link, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	item, ok := repository.index[id]
	if !ok {
		var zero model.Link
		return zero, errors.New("not found")
	}

	return item, nil
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

	indices, ok := repository.userIndex[userID]
	if !ok {
		return []model.Link{}, nil
	}

	result := make([]model.Link, 0, len(indices))
	for _, idx := range indices {
		item := repository.data[idx]
		if !item.IsDeleted {
			result = append(result, item)
		}
	}

	return result, nil
}

func (repository *InMemoryRepository) SoftDeleteByIDs(ids []string, userID string) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	idSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}

	for i := range repository.data {
		if repository.data[i].CreatedBy == userID && idSet[repository.data[i].ID] {
			repository.data[i].IsDeleted = true
			repository.index[repository.data[i].ID] = repository.data[i]
		}
	}

	return nil
}
