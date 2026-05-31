package handler

import (
	"encoding/json"
	"net/http"
	"shorturl/internal/config"
	"shorturl/internal/context"
	"shorturl/internal/dto"
	"shorturl/internal/service"

	"go.uber.org/zap"
)

// ListUrls creates and returns an HTTP handler for retrieving all URLs created by the authenticated user.
//
// The handler accepts GET requests and returns a JSON array of all original URLs and their
// corresponding short URLs for the authenticated user. If the user has no URLs, it returns
// 204 No Content.
//
// Request format:
//   GET /api/user/urls
//
// Response format (200 OK):
//   Content-Type: application/json
//   Body: [
//     {
//       "original_url": "https://example.com",
//       "short_url": "http://localhost:8080/abc123"
//     },
//     {
//       "original_url": "https://example2.com",
//       "short_url": "http://localhost:8080/def456"
//     }
//   ]
//
// Response format (204 No Content):
//   User has no URLs created
//
// Error responses:
//   Error (500 Internal Server Error): Internal server error
//
// Parameters:
//   - linkService: LinkServiceInterface for retrieving user links
//   - logger: Logger for logging errors and information
//   - config: Configuration containing BaseURL for building short URLs
//
// Returns:
//   - http.HandlerFunc: Handler function for listing user URLs
func ListUrls(linkService service.LinkServiceInterface, logger *zap.Logger, config config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := context.UserID(r.Context())
		if !ok {
			logger.Error("User id not found in context")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		logger.Info("requesting links for user", zap.String("userID", userID))
		links, err := linkService.GetAllByUserID(userID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
			return
		}

		if len(links) == 0 {
			logger.Info("No links for user")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		preparedResponseLinks := make([]dto.ResponseListItem, len(links))
		for index, value := range links {
			preparedResponseLink := dto.ResponseListItem{
				OriginalURL: value.URL,
				ShortURL:    config.BaseURL + "/" + value.ID,
			}
			preparedResponseLinks[index] = preparedResponseLink
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(preparedResponseLinks); err != nil {
			logger.Error("failed to write response", zap.Error(err))
		}
	}
}
