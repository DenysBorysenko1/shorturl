package handler

import (
	"net/http"
	"shorturl/internal/service"

	"github.com/go-chi/chi/v5"
)

func Retrieve(service service.LinkServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.NotFound(w, r)
			return
		}

		url, err := service.Get(id)

		if err != nil {
			http.Error(w, "Error while retrieving", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
