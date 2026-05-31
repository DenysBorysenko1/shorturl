package handler

import (
	"errors"
	"net/http"
	"shorturl/internal/audit"
	"shorturl/internal/service"
	"time"

	"github.com/go-chi/chi/v5"
)

// Retrieve creates and returns an HTTP handler for retrieving and redirecting to original URLs.
//
// The handler accepts GET requests with a short URL ID as a path parameter and redirects
// the client to the original URL. If the link is deleted, it returns 410 Gone. If the link
// doesn't exist, it returns 404 Not Found. The handler also tracks retrieval events via
// the audit broadcaster.
//
// Request format:
//   GET /{id}
//
// Response formats:
//   Success (307 Temporary Redirect):
//     Location: https://example.com (the original URL)
//
//   Error (404 Not Found): Link not found
//   Error (410 Gone): Link has been deleted
//   Error (500 Internal Server Error): Internal server error
//
// Parameters:
//   - svc: LinkServiceInterface for retrieving links
//   - broadcaster: Audit event broadcaster for tracking link retrieval events
//
// Returns:
//   - http.HandlerFunc: Handler function for URL retrieval and redirection
func Retrieve(svc service.LinkServiceInterface, broadcaster *audit.Broadcaster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.NotFound(w, r)
			return
		}

		url, err := svc.Get(id)

		if err != nil {
			if errors.Is(err, service.ErrLinkDeleted) {
				w.WriteHeader(http.StatusGone)
				return
			}
			if errors.Is(err, service.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "Error while retrieving", http.StatusInternalServerError)
			return
		}

		broadcaster.Notify(audit.Event{
			TS:     time.Now().Unix(),
			Action: "retrieve",
			UserID: "",
			URL:    url,
		})

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
