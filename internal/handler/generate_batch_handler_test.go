package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"shorturl/internal/config"
	"shorturl/internal/dto"
	"shorturl/internal/handler"
	"shorturl/internal/model"
	"shorturl/internal/service/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateBatch(t *testing.T) {

	config.Load()

	type want struct {
		status        int
		contentType   string
		response      []dto.ResponseBatchItem
		errorResponse *handler.ErrorResponse
	}

	tests := []struct {
		name           string
		method         string
		path           string
		request        []dto.RequestBatchItem
		createManyFunc func(links []model.Link) error
		want           want
	}{
		{
			name:   "positive",
			method: "POST",
			path:   "/api/shorten/batch",
			request: []dto.RequestBatchItem{
				{
					CorrelationID: "f0i1r2s3t",
					OriginalURL:   "https://first.com",
				},
				{
					CorrelationID: "s1e2c3o4n5d",
					OriginalURL:   "https://first.com",
				},
			},
			want: want{
				status:      http.StatusCreated,
				contentType: "application/json",
				response: []dto.ResponseBatchItem{
					{
						CorrelationID: "f0i1r2s3t",
						ShortURL:      config.Cfg.BaseURL + "/f0i1r2s3t",
					},
					{
						CorrelationID: "s1e2c3o4n5d",
						ShortURL:      config.Cfg.BaseURL + "/s1e2c3o4n5d",
					},
				},
			},
		},
		{
			name:    "empty batch",
			method:  "POST",
			path:    "/api/shorten/batch",
			request: []dto.RequestBatchItem{},
			want: want{
				status:      http.StatusBadRequest,
				contentType: "application/json",
				errorResponse: &handler.ErrorResponse{
					Error: "Empty data",
				},
			},
		},
		{
			name:   "service error",
			method: "POST",
			path:   "/api/shorten/batch",
			request: []dto.RequestBatchItem{
				{
					CorrelationID: "f0i1r2s3t",
					OriginalURL:   "https://first.com",
				},
			},
			createManyFunc: func(links []model.Link) error {
				return errors.New("error")
			},
			want: want{
				status:      http.StatusInternalServerError,
				contentType: "application/json",
				errorResponse: &handler.ErrorResponse{
					Error: "Error while creating",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &mocks.MockLinkService{
				CreateManyFunc: func(links []model.Link) error {
					if test.createManyFunc != nil {
						return test.createManyFunc(links)
					}

					return nil
				},
			}

			jsonRequest, err := json.Marshal(test.request)
			require.NoError(t, err)

			request := httptest.NewRequest(test.method, test.path, bytes.NewReader(jsonRequest))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(handler.GenerateBatch(service, config.Cfg))

			h(w, request)

			result := w.Result()
			defer result.Body.Close()
			data, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			assert.Equal(t, test.want.status, result.StatusCode)
			assert.Contains(t, result.Header.Get("Content-Type"), test.want.contentType)

			if test.want.response != nil {
				var response []dto.ResponseBatchItem
				err = json.Unmarshal(data, &response)
				require.NoError(t, err)

				assert.Equal(t, test.want.response, response)
			}

			if test.want.errorResponse != nil {
				var errorResponse handler.ErrorResponse
				err = json.Unmarshal(data, &errorResponse)
				require.NoError(t, err)

				assert.Equal(t, *test.want.errorResponse, errorResponse)
			}

		})
	}
}
