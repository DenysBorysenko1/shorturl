package mocks

import "shorturl/internal/model"

type MockLinkService struct {
	CreateFunc     func(url string) (string, error)
	CreateManyFunc func(links []model.Link) error
	GetFunc        func(url string) (string, error)
}

func (linkService *MockLinkService) Create(url string) (string, error) {
	if linkService.CreateFunc != nil {
		return linkService.CreateFunc(url)
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
