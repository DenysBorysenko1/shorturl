package handler

import (
	"net/http"
	"shorturl/internal/service"
)

func Retrieve(service *service.LinkService) http.HandlerFunc {
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
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
