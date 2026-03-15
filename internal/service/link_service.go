package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"shorturl/internal/logger"
	"shorturl/internal/model"
	"shorturl/internal/repository"

	"go.uber.org/zap"
)

type LinkServiceInterface interface {
	Create(url string) (string, error)
	Get(id string) (string, error)
}

type LinkService struct {
	repository repository.Repository[model.Link]
	logger     zap.Logger
}

func NewLinkService(repository repository.Repository[model.Link], logger zap.Logger) *LinkService {
	return &LinkService{repository: repository, logger: logger}
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
		logger.Log.Error(err.Error())

		return "", fmt.Errorf("create link: %w", err)
	}

	return id, nil
}

func (linkService *LinkService) Get(id string) (string, error) {
	data, err := linkService.repository.GetByID(id)
	if err != nil {
		logger.Log.Error(err.Error())

		return "", errors.New("not found")
	}

	return data.URL, nil
}
