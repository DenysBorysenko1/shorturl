package handler

import (
	"net/http"
)

func Retrieve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	http.Redirect(w, r, "https://yandex.ru/", http.StatusTemporaryRedirect)
}