package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"shorturl/internal/model"
	"shorturl/internal/repository"
)

type LinkService struct {
	repository repository.Repository[model.Link]
}

func NewLinkService(repository repository.Repository[model.Link]) *LinkService {
	return &LinkService{repository: repository}
}

func (linkService *LinkService) Create(url string) (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	id := hex.EncodeToString(bytes)

	link := model.Link{
		BaseEntity: model.BaseEntity{ID: id},
		URL:        url,
	}
	if err := linkService.repository.Create(link); err != nil {
		return "", fmt.Errorf("create link: %w", err)
	}

	return id, nil
}

func (linkService *LinkService) Get(id string) (string, error) {
	data, err := linkService.repository.GetById(id)
	if err != nil {
		fmt.Println(err)
		return "", errors.New("not found")
	}

	return data.URL, nil
}
