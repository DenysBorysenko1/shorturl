package mocks

type MockLinkService struct {
	CreateFunc func(url string) (string, error)
	GetFunc    func(url string) (string, error)
}

func (linkService *MockLinkService) Create(url string) (string, error) {
	if linkService.CreateFunc != nil {
		return linkService.CreateFunc(url)
	}

	return "", nil
}

func (linkService *MockLinkService) Get(id string) (string, error) {
	if linkService.GetFunc != nil {
		return linkService.GetFunc(id)
	}

	return "", nil
}
