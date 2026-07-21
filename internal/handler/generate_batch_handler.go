package handler

import (
	"encoding/json"
	"net/http"
	"shorturl/internal/config"
	"shorturl/internal/context"
	"shorturl/internal/dto"
	"shorturl/internal/model"
	"shorturl/internal/service"

	"go.uber.org/zap"
)

// GenerateBatch creates and returns an HTTP handler for batch shortening of URLs.
//
// The handler accepts POST requests with a JSON array of items, each containing a
// correlation ID and original URL. It returns a JSON array with correlation IDs and
// short URLs. All URLs are associated with the authenticated user ID from the request
// context.
//
// Request format:
//
//	POST /api/shorten/batch
//	Content-Type: application/json
//	Body: [
//	  {
//	    "correlation_id": "id1",
//	    "original_url": "https://example1.com"
//	  },
//	  {
//	    "correlation_id": "id2",
//	    "original_url": "https://example2.com"
//	  }
//	]
//
// Response format (201 Created):
//
//	Content-Type: application/json
//	Body: [
//	  {
//	    "correlation_id": "id1",
//	    "short_url": "http://localhost:8080/id1"
//	  },
//	  {
//	    "correlation_id": "id2",
//	    "short_url": "http://localhost:8080/id2"
//	  }
//	]
//
// Error responses:
//
//	Error (400 Bad Request): Invalid JSON or empty data
//	Error (405 Method Not Allowed): Only POST method is allowed
//	Error (500 Internal Server Error): Internal server error
//
// Parameters:
//   - svc: LinkServiceInterface for batch creating links
//   - logger: Logger for logging errors and information
//   - config: Configuration containing BaseURL for building short URLs
//
// Returns:
//   - http.HandlerFunc: Handler function for batch URL shortening
func GenerateBatch(svc service.LinkServiceInterface, logger *zap.Logger, config config.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {

		userID, ok := context.UserID(req.Context())
		if !ok {
			logger.Error("User id not found in context")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var inputItems []dto.RequestBatchItem
		if err := json.NewDecoder(req.Body).Decode(&inputItems); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			if encErr := json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()}); encErr != nil {
				logger.Error("failed to encode error response", zap.Error(encErr))
			}
			return
		}

		if len(inputItems) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if encErr := json.NewEncoder(w).Encode(ErrorResponse{Error: "Empty data"}); encErr != nil {
				logger.Error("failed to encode error response", zap.Error(encErr))
			}

			return
		}

		preparedInputLinks := make([]model.Link, len(inputItems))

		for index, value := range inputItems {
			preparedInputLink := model.Link{
				BaseEntity: model.BaseEntity{ID: value.CorrelationID},
				URL:        value.OriginalURL,
				CreatedBy:  userID,
			}
			preparedInputLinks[index] = preparedInputLink
		}

		err := svc.CreateMany(preparedInputLinks)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			if encErr := json.NewEncoder(w).Encode(ErrorResponse{Error: "Error while creating"}); encErr != nil {
				logger.Error("failed to encode error response", zap.Error(encErr))
			}
			return
		}

		preparedResponseLinks := make([]dto.ResponseBatchItem, len(inputItems))
		for index, value := range inputItems {
			preparedResponseLink := dto.ResponseBatchItem{
				CorrelationID: value.CorrelationID,
				ShortURL:      config.BaseURL + "/" + value.CorrelationID,
			}
			preparedResponseLinks[index] = preparedResponseLink
		}

		resp, err := json.Marshal(preparedResponseLinks)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if _, writeErr := w.Write(resp); writeErr != nil {
			logger.Error("failed to write response", zap.Error(writeErr))
		}
	}
}
