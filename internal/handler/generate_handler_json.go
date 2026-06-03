package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"shorturl/internal/audit"
	"shorturl/internal/config"
	"shorturl/internal/context"
	"shorturl/internal/service"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Input represents the JSON request body for creating a short URL.
type Input struct {
	URL string `json:"url"` // The original URL to shorten
}

// Response represents the JSON response containing the short URL.
type Response struct {
	Result string `json:"result"` // The generated short URL
}

// ErrorResponse represents an error response in JSON format.
type ErrorResponse struct {
	Error string `json:"error"` // Error message describing what went wrong
}

// GenerateJSON creates and returns an HTTP handler for shortening URLs using JSON format.
//
// The handler accepts POST requests with a JSON body containing the original URL
// and returns a JSON response with the shortened URL. If the URL already exists,
// it returns the existing short URL with a 409 Conflict status. The URL is associated
// with the authenticated user ID from the request context.
//
// Request format:
//
//	POST /api/shorten
//	Content-Type: application/json
//	Body: {"url": "https://example.com"}
//
// Response formats:
//
//	Success (201 Created):
//	  Content-Type: application/json
//	  Body: {"result": "http://localhost:8080/abc123"}
//
//	Conflict (409 Conflict) - URL already exists:
//	  Content-Type: application/json
//	  Body: {"result": "http://localhost:8080/abc123"}
//
//	Error (400 Bad Request):
//	  Content-Type: application/json
//	  Body: {"error": "error message"}
//
//	Error (405 Method Not Allowed): Only POST method is allowed
//	Error (500 Internal Server Error): Internal server error
//
// Parameters:
//   - svc: LinkServiceInterface for creating and managing links
//   - logger: Logger for logging errors and information
//   - config: Configuration containing BaseURL for building short URLs
//   - broadcaster: Audit event broadcaster for tracking link creation events
//
// Returns:
//   - http.HandlerFunc: Handler function for JSON-based URL shortening
func GenerateJSON(svc service.LinkServiceInterface, logger *zap.Logger, config config.Config, broadcaster *audit.Broadcaster) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		userID, ok := context.UserID(req.Context())
		if !ok {
			logger.Error("User id not found in context")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var input Input
		if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
			return
		}

		url := strings.TrimSpace(string(input.URL))
		if url == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "URL is required"})

			return
		}

		id, err := svc.Create(url, userID)
		if err != nil {
			var conflictErr *service.URLAlreadyExistsError
			if errors.As(err, &conflictErr) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(Response{
					Result: config.BaseURL + "/" + conflictErr.ID,
				})
				broadcaster.Notify(audit.Event{
					TS:     time.Now().Unix(),
					Action: "shorten",
					UserID: userID,
					URL:    url,
				})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Error while creating"})
			return
		}

		resultURL := config.BaseURL + "/" + id

		resp, err := json.Marshal(Response{
			Result: resultURL,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)

		broadcaster.Notify(audit.Event{
			TS:     time.Now().Unix(),
			Action: "shorten",
			UserID: userID,
			URL:    url,
		})
	}
}
