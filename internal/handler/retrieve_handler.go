package handler

import (
	"errors"
	"net/http"
	"shorturl/internal/audit"
	"shorturl/internal/service"
	"time"

	"github.com/go-chi/chi/v5"
)

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
