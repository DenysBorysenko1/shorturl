package handler

import (
	"encoding/json"
	"net/http"
	"shorturl/internal/config"
	"shorturl/internal/context"
	"shorturl/internal/dto"
	"shorturl/internal/model"
	"shorturl/internal/service"

	"go.uber.org/zap"
)

func GenerateBatch(svc service.LinkServiceInterface, logger *zap.Logger, config config.Config) http.HandlerFunc {

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

		var inputItems []dto.RequestBatchItem
		if err := json.NewDecoder(req.Body).Decode(&inputItems); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
			return
		}

		if len(inputItems) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Empty data"})

			return
		}

		preparedInputLinks := make([]model.Link, len(inputItems))

		for index, value := range inputItems {
			preparedInputLink := model.Link{
				BaseEntity: model.BaseEntity{ID: value.CorrelationID},
				URL:        value.OriginalURL,
				CreatedBy:  userID,
			}
			preparedInputLinks[index] = preparedInputLink
		}

		err := svc.CreateMany(preparedInputLinks)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Error while creating"})
			return
		}

		preparedResponseLinks := make([]dto.ResponseBatchItem, len(inputItems))
		for index, value := range inputItems {
			preparedResponseLink := dto.ResponseBatchItem{
				CorrelationID: value.CorrelationID,
				ShortURL:      config.BaseURL + "/" + value.CorrelationID,
			}
			preparedResponseLinks[index] = preparedResponseLink
		}

		resp, err := json.Marshal(preparedResponseLinks)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
	}
}
