package handler

import (
	"net/http"
	"shorturl/internal/service"
)

func Retrieve(service service.LinkServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET", http.StatusMethodNotAllowed)
			return
		}
		id := r.PathValue("id")
		if id == "" {
			http.NotFound(w, r)
			return
		}

		url, err := service.Get(id)

		if err != nil {
			http.Error(w, "Error while creating", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
