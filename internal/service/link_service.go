package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"shorturl/internal/config"
	"shorturl/internal/logger"
	"shorturl/internal/model"
	"shorturl/internal/repository"

	"go.uber.org/zap"
)

type LinkServiceInterface interface {
	CreateMany(links []model.Link) error
	Create(url string) (string, error)
	Get(id string) (string, error)
}

type LinkService struct {
	repository repository.Repository[model.Link]
	logger     zap.Logger
	config     config.Config
}

func NewLinkService(repository repository.Repository[model.Link], logger zap.Logger) *LinkService {
	return &LinkService{repository: repository, logger: logger}
}

func (linkService *LinkService) CreateMany(links []model.Link) error {
	res, err := linkService.repository.CreateMany(links)
	linkService.logger.Info("Batch links was added", zap.Int("count", res))

	if err != nil {
		linkService.logger.Info("Error while adding batch", zap.Error(err))
		return err
	}

	return nil
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
