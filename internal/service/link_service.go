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
	"sync"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type LinkServiceInterface interface {
	GetByURL(url string) (model.Link, error)
	CreateMany(links []model.Link) error
	GetAllByUserID(userID string) ([]model.Link, error)
	Create(url string, userID string) (string, error)
	Get(id string) (string, error)
	EnqueueDelete(ids []string, userID string) error
}

type LinkService struct {
	repository repository.Repository[model.Link]
	logger     zap.Logger
	config     config.Config

	deleteOnce sync.Once
	deleteCh   chan deleteJob
}

var (
	ErrNotFound    = errors.New("not found")
	ErrLinkDeleted = errors.New("link deleted")
)

type deleteJob struct {
	userID string
	ids    []string
}

type URLAlreadyExistsError struct {
	ID string
}

func (e *URLAlreadyExistsError) Error() string {
	return "url already exists"
}

func NewLinkService(repository repository.Repository[model.Link], logger zap.Logger) *LinkService {
	svc := &LinkService{
		repository: repository,
		logger:     logger,
		deleteCh:   make(chan deleteJob, 1024),
	}
	svc.startDeleteWorker()
	return svc
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

func (linkService *LinkService) Create(url string, userID string) (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	id := hex.EncodeToString(bytes)

	link := model.Link{
		BaseEntity: model.BaseEntity{ID: id},
		URL:        url,
		CreatedBy:  userID,
	}
	if err := linkService.repository.Create(link); err != nil {
		logger.Log.Error(err.Error())

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			existedURL, getErr := linkService.repository.GetByURL(url)
			if getErr != nil {
				return "", fmt.Errorf("get existed url: %w", getErr)
			}

			return "", &URLAlreadyExistsError{ID: existedURL.ID}
		}

		return "", fmt.Errorf("create link: %w", err)
	}

	return id, nil
}

func (linkService *LinkService) Get(id string) (string, error) {
	data, err := linkService.repository.GetByID(id)
	if err != nil {
		logger.Log.Error(err.Error())

		return "", ErrNotFound
	}

	if data.IsDeleted {
		return "", ErrLinkDeleted
	}

	return data.URL, nil
}

func (linkService *LinkService) GetByURL(url string) (model.Link, error) {
	link, err := linkService.repository.GetByURL(url)
	if err != nil {
		var zero model.Link
		return zero, err
	}

	return link, nil
}

func (linkService *LinkService) GetAllByUserID(userID string) ([]model.Link, error) {
	links, err := linkService.repository.GetAllByUserID(userID)
	if err != nil {
		linkService.logger.Info("Error while retrieving all links for", zap.Error(err))

		return nil, err
	}

	return links, nil
}

func (linkService *LinkService) EnqueueDelete(ids []string, userID string) error {
	if len(ids) == 0 {
		return nil
	}
	linkService.startDeleteWorker()
	linkService.deleteCh <- deleteJob{userID: userID, ids: ids}
	return nil
}

func (linkService *LinkService) startDeleteWorker() {
	linkService.deleteOnce.Do(func() {
		go linkService.deleteWorker()
	})
}

func (linkService *LinkService) flush(buf map[string][]string) {
	for userID, ids := range buf {
		if len(ids) == 0 {
			continue
		}

		if err := linkService.repository.SoftDeleteByIDs(ids, userID); err != nil {
			linkService.logger.Info("failed to soft-delete links", zap.Error(err))
		}
	}
}

func (linkService *LinkService) deleteWorker() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	buf := make(map[string][]string)

	for {
		select {
		case job := <-linkService.deleteCh:
			if len(job.ids) == 0 || job.userID == "" {
				continue
			}
			buf[job.userID] = append(buf[job.userID], job.ids...)

		case <-ticker.C:
			linkService.flush(buf)
			buf = make(map[string][]string)
		}
	}
}
