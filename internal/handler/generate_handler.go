package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"shorturl/internal/config"
	"shorturl/internal/context"
	"shorturl/internal/service"
	"strings"

	"go.uber.org/zap"
)

func Generate(svc service.LinkServiceInterface, logger *zap.Logger, config config.Config) http.HandlerFunc {

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
				return
			}

			http.Error(w, "Error while creating", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(config.BaseURL + "/" + id))
	}

}
