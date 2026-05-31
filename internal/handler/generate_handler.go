package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"shorturl/internal/audit"
	"shorturl/internal/config"
	"shorturl/internal/context"
	"shorturl/internal/service"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Generate creates and returns an HTTP handler for shortening URLs.
//
// The handler accepts POST requests with a URL in the request body (plain text format)
// and returns a shortened URL. If the URL already exists, it returns the existing short URL
// with a 409 Conflict status. The URL is associated with the authenticated user ID from
// the request context.
//
// Request format:
//   POST /shorten
//   Content-Type: text/plain
//   Body: https://example.com
//
// Response formats:
//   Success (201 Created):
//     Content-Type: text/plain
//     Body: http://localhost:8080/abc123
//
//   Conflict (409 Conflict) - URL already exists:
//     Content-Type: text/plain
//     Body: http://localhost:8080/abc123
//
//   Error (400 Bad Request): URL is required
//   Error (405 Method Not Allowed): Only POST method is allowed
//   Error (500 Internal Server Error): Internal server error
//
// Parameters:
//   - svc: LinkServiceInterface for creating and managing links
//   - logger: Logger for logging errors and information
//   - config: Configuration containing BaseURL for building short URLs
//   - broadcaster: Audit event broadcaster for tracking link creation events
//
// Returns:
//   - http.HandlerFunc: Handler function for URL shortening
func Generate(svc service.LinkServiceInterface, logger *zap.Logger, config config.Config, broadcaster *audit.Broadcaster) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := context.UserID(r.Context())
		if !ok {
			logger.Error("User id not found in context")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Only POST", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 2048))
		if err != nil {
			fmt.Println(err)
		}

		defer r.Body.Close()
		url := strings.TrimSpace(string(body))
		if url == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		id, err := svc.Create(url, userID)
		if err != nil {
			var conflictErr *service.URLAlreadyExistsError
			if errors.As(err, &conflictErr) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(config.BaseURL + "/" + conflictErr.ID))
				broadcaster.Notify(audit.Event{
					TS:     time.Now().Unix(),
					Action: "shorten",
					UserID: userID,
					URL:    url,
				})
				return
			}

			http.Error(w, "Error while creating", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(config.BaseURL + "/" + id))

		broadcaster.Notify(audit.Event{
			TS:     time.Now().Unix(),
			Action: "shorten",
			UserID: userID,
			URL:    url,
		})
	}

}
