package handler

import (
	"encoding/json"
	"net/http"
	"shorturl/internal/config"
	"shorturl/internal/context"
	"shorturl/internal/service"

	"go.uber.org/zap"
)

// RemoveListUrls creates and returns an HTTP handler for deleting multiple URLs.
//
// The handler accepts DELETE requests with a JSON array of short URL IDs to delete.
// The deletion is performed asynchronously in the background. Only the authenticated
// user's URLs can be deleted. The handler returns 202 Accepted immediately when the
// deletion task is enqueued.
//
// Request format:
//
//	DELETE /api/user/urls
//	Content-Type: application/json
//	Body: ["abc123", "def456", "ghi789"]
//
// Response format (202 Accepted):
//
//	Deletion task has been enqueued for processing
//
// Error responses:
//
//	Error (400 Bad Request): Invalid JSON format
//	Error (500 Internal Server Error): Internal server error
//
// Parameters:
//   - linkService: LinkServiceInterface for deleting links
//   - logger: Logger for logging errors and information
//   - config: Configuration object
//
// Returns:
//   - http.HandlerFunc: Handler function for deleting multiple URLs
func RemoveListUrls(linkService service.LinkServiceInterface, logger *zap.Logger, config config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := context.UserID(r.Context())
		if !ok {
			logger.Error("User id not found in context")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var ids []string
		if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
			return
		}

		_ = linkService.EnqueueDelete(ids, userID)
		w.WriteHeader(http.StatusAccepted)

	}
}
