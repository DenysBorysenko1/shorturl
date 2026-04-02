package mocks

import "shorturl/internal/model"

type MockLinkService struct {
	GetByURLFunc       func(url string) (model.Link, error)
	CreateFunc         func(url string, userID string) (string, error)
	CreateManyFunc     func(links []model.Link) error
	GetFunc            func(url string) (string, error)
	GetAllByUserIDFunc func(userID string) ([]model.Link, error)
	EnqueueDeleteFunc  func(ids []string, userID string) error
}

func (linkService *MockLinkService) Create(url string, userID string) (string, error) {
	if linkService.CreateFunc != nil {
		return linkService.CreateFunc(url, userID)
	}

	return "", nil
}

func (linkService *MockLinkService) CreateMany(links []model.Link) error {
	if linkService.CreateManyFunc != nil {
		return linkService.CreateManyFunc(links)
	}

	return nil
}

func (linkService *MockLinkService) Get(id string) (string, error) {
	if linkService.GetFunc != nil {
		return linkService.GetFunc(id)
	}

	return "", nil
}

func (linkService *MockLinkService) GetByURL(url string) (model.Link, error) {
	if linkService.GetByURLFunc != nil {
		return linkService.GetByURLFunc(url)
	}

	var zero = model.Link{}
	return zero, nil
}

func (linkService *MockLinkService) GetAllByUserID(userID string) ([]model.Link, error) {
	if linkService.GetAllByUserIDFunc != nil {
		return linkService.GetAllByUserIDFunc(userID)
	}

	return nil, nil
}

func (linkService *MockLinkService) EnqueueDelete(ids []string, userID string) error {
	if linkService.EnqueueDeleteFunc != nil {
		return linkService.EnqueueDeleteFunc(ids, userID)
	}
	return nil
}
