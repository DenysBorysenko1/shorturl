package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"shorturl/internal/config"
	"shorturl/internal/context"
	"shorturl/internal/service"
	"strings"

	"go.uber.org/zap"
)

type Input struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func GenerateJSON(svc service.LinkServiceInterface, logger *zap.Logger, config config.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		userID, ok := context.UserID(req.Context())
		if !ok {
			logger.Error("User id not found in context")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if req.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)

			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Only POST"})
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
	}
}
