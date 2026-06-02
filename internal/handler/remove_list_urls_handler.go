package handler

import (
	"encoding/json"
	"net/http"
	"shorturl/internal/config"
	"shorturl/internal/context"
	"shorturl/internal/service"

	"go.uber.org/zap"
)

func RemoveListUrls(linkService service.LinkServiceInterface, logger *zap.Logger, config config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := context.UserID(r.Context())
		if !ok {
			logger.Error("User id not found in context")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var ids []string
		if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
			return
		}

		_ = linkService.EnqueueDelete(ids, userID)
		w.WriteHeader(http.StatusAccepted)

	}
}
