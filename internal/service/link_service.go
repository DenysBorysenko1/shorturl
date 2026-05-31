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

// LinkServiceInterface defines the contract for link service operations.
//
// This interface provides methods for creating, retrieving, and managing short links.
// Implementations should handle business logic for URL shortening, user associations,
// and asynchronous deletion operations.
type LinkServiceInterface interface {
	// GetByURL retrieves a link by its original URL.
	GetByURL(url string) (model.Link, error)
	// CreateMany creates multiple links in a single batch operation.
	CreateMany(links []model.Link) error
	// GetAllByUserID retrieves all links created by a specific user.
	GetAllByUserID(userID string) ([]model.Link, error)
	// Create creates a new short link for the given URL and associates it with a user.
	Create(url string, userID string) (string, error)
	// Get retrieves the original URL for a given short link ID.
	Get(id string) (string, error)
	// EnqueueDelete enqueues multiple link IDs for asynchronous deletion.
	EnqueueDelete(ids []string, userID string) error
}

// LinkService implements LinkServiceInterface with in-memory storage,
// batch deletion support, and audit logging.
type LinkService struct {
	repository repository.Repository[model.Link]
	logger     zap.Logger
	config     config.Config

	deleteOnce sync.Once
	deleteCh   chan deleteJob
}

var (
	// ErrNotFound is returned when a link cannot be found.
	ErrNotFound    = errors.New("not found")
	// ErrLinkDeleted is returned when attempting to retrieve a deleted link.
	ErrLinkDeleted = errors.New("link deleted")
)

// deleteJob represents a batch deletion job for async processing.
type deleteJob struct {
	userID string
	ids    []string
}

// URLAlreadyExistsError is returned when attempting to create a short link
// for a URL that already exists in the system.
type URLAlreadyExistsError struct {
	ID string // The existing short link ID
}

// Error implements the error interface.
func (e *URLAlreadyExistsError) Error() string {
	return "url already exists"
}

// NewLinkService creates a new LinkService instance with the given repository and logger.
// It also starts the background delete worker for handling asynchronous deletions.
//
// Parameters:
//   - repository: Repository instance for storing and retrieving links
//   - logger: Logger instance for logging operations
//
// Returns:
//   - *LinkService: A new LinkService instance
func NewLinkService(repository repository.Repository[model.Link], logger zap.Logger) *LinkService {
	svc := &LinkService{
		repository: repository,
		logger:     logger,
		deleteCh:   make(chan deleteJob, 1024),
	}
	svc.startDeleteWorker()
	return svc
}

// CreateMany creates multiple links in a single batch operation.
//
// This method stores all provided links in the repository and logs the number
// of successfully created links. If an error occurs, it logs the error and returns it.
//
// Parameters:
//   - links: Slice of Link objects to create
//
// Returns:
//   - error: Error if the batch creation fails, nil otherwise
func (linkService *LinkService) CreateMany(links []model.Link) error {
	res, err := linkService.repository.CreateMany(links)
	linkService.logger.Info("Batch links was added", zap.Int("count", res))

	if err != nil {
		linkService.logger.Info("Error while adding batch", zap.Error(err))
		return err
	}

	return nil
}

// Create creates a new short link for the given URL and associates it with a user.
//
// This method generates a random 8-character hexadecimal ID for the short link and
// attempts to store it. If the URL already exists in the system, it returns an
// URLAlreadyExistsError with the existing short link ID.
//
// Parameters:
//   - url: The original URL to shorten
//   - userID: The ID of the user creating this link
//
// Returns:
//   - string: The generated short link ID
//   - error: URLAlreadyExistsError if URL exists, or other error if creation fails
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

// Get retrieves the original URL for a given short link ID.
//
// This method looks up the link by its ID and returns the original URL. If the link
// is marked as deleted, it returns ErrLinkDeleted. If the link is not found, it returns
// ErrNotFound.
//
// Parameters:
//   - id: The short link ID to look up
//
// Returns:
//   - string: The original URL if found
//   - error: ErrLinkDeleted if link is deleted, ErrNotFound if not found, or other error
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

// GetByURL retrieves a link by its original URL.
//
// This method looks up a link by the original URL and returns the complete Link object
// including its ID and metadata. This is useful for checking if a URL has already been
// shortened.
//
// Parameters:
//   - url: The original URL to look up
//
// Returns:
//   - model.Link: The link object if found
//   - error: Error if the link cannot be retrieved
func (linkService *LinkService) GetByURL(url string) (model.Link, error) {
	link, err := linkService.repository.GetByURL(url)
	if err != nil {
		var zero model.Link
		return zero, err
	}

	return link, nil
}

// GetAllByUserID retrieves all links created by a specific user.
//
// This method returns a slice of all Link objects associated with the given user ID.
// The links are returned in the order they are stored by the repository.
//
// Parameters:
//   - userID: The ID of the user whose links to retrieve
//
// Returns:
//   - []model.Link: Slice of links created by the user
//   - error: Error if retrieval fails
func (linkService *LinkService) GetAllByUserID(userID string) ([]model.Link, error) {
	links, err := linkService.repository.GetAllByUserID(userID)
	if err != nil {
		linkService.logger.Info("Error while retrieving all links for", zap.Error(err))

		return nil, err
	}

	return links, nil
}

// EnqueueDelete enqueues multiple link IDs for asynchronous deletion.
//
// This method adds the specified link IDs to a deletion queue that is processed by a
// background worker. The deletion is performed asynchronously, typically after a 10-second
// batching period. Only empty ID slices are no-ops.
//
// Parameters:
//   - ids: Slice of link IDs to delete
//   - userID: The ID of the user requesting the deletion
//
// Returns:
//   - error: Always nil (deletion is queued for background processing)
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
