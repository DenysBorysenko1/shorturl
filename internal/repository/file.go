package repository

import (
	"encoding/json"
	"errors"
	"os"
	"shorturl/internal/model"
	"slices"
	"sync"
)

type FileRepository struct {
	filename string
	mu       sync.RWMutex
}

func NewFileRepository[T Entity](filename string) (*FileRepository, error) {
	repo := &FileRepository{
		filename: filename,
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if err := os.WriteFile(filename, []byte("[]"), 0666); err != nil {
			return nil, err
		}
	}

	return repo, nil
}

func (r *FileRepository) CreateMany(entities []model.Link) (int, error) {
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

func (r *FileRepository) loadAll() ([]model.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := os.ReadFile(r.filename)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return []model.Link{}, nil
	}

	var items []model.Link
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *FileRepository) saveAll(items []model.Link) error {
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

func (r *FileRepository) Create(entity model.Link) error {
	items, err := r.loadAll()
	if err != nil {
		return err
	}

	items = append(items, entity)

	return r.saveAll(items)
}

func (r *FileRepository) GetByID(id string) (model.Link, error) {
	items, err := r.loadAll()
	if err != nil {
		var zero model.Link
		return zero, err
	}

	for _, item := range items {
		if item.GetID() == id {
			return item, nil
		}
	}

	var zero model.Link
	return zero, errors.New("not found")
}

func (r *FileRepository) GetByURL(url string) (model.Link, error) {
	items, err := r.loadAll()
	if err != nil {
		var zero model.Link
		return zero, err
	}

	for _, item := range items {
		if item.GetURL() == url {
			return item, nil
		}
	}

	var zero model.Link
	return zero, errors.New("not found")
}

func (r *FileRepository) GetAllByUserID(userID string) ([]model.Link, error) {
	items, err := r.loadAll()
	if err != nil {
		return nil, err
	}

	var result []model.Link
	for _, item := range items {
		if item.GetCreatedBy() == userID && !item.IsDeleted {
			result = append(result, item)
		}
	}

	return result, nil
}

func (r *FileRepository) SoftDeleteByIDs(ids []string, userID string) error {
	items, err := r.loadAll()
	if err != nil {
		return err
	}

	for i := range items {
		if items[i].CreatedBy == userID && slices.Contains(ids, items[i].ID) {
			items[i].IsDeleted = true
		}
	}

	return r.saveAll(items)
}
