package handler

import (
	"fmt"
	"io"
	"net/http"
	"shorturl/internal/service"
	"strings"
)

func Generate(svc service.LinkServiceInterface, baseURL string) http.HandlerFunc {
	
	return func(w http.ResponseWriter, r *http.Request) {
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

		id, err := svc.Create(url)
		if err != nil {
			http.Error(w, "Error while creating", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(baseURL + "/" + id))
	}

}
