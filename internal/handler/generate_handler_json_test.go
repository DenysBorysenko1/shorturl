package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"shorturl/internal/audit"
	"shorturl/internal/config"
	appctx "shorturl/internal/context"
	"shorturl/internal/handler"
	"shorturl/internal/model"
	"shorturl/internal/service"
	"shorturl/internal/service/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGenerateJSON(t *testing.T) {
	config.Load()

	type want struct {
		status        int
		contentType   string
		response      *handler.Response
		errorResponse *handler.ErrorResponse
	}

	tests := []struct {
		name       string
		id         string
		method     string
		path       string
		request    handler.Input
		createFunc func(url string, userID string) (string, error)
		want       want
	}{
		{
			name:   "positive",
			method: "POST",
			path:   "/api/shorten",
			id:     "123",
			request: handler.Input{
				URL: "https://yandex.ru",
			},
			want: want{
				status:      http.StatusCreated,
				contentType: "application/json",
				response: &handler.Response{
					Result: config.Cfg.BaseURL + "/123",
				},
			},
		},
		{
			name:   "not pass param",
			method: "POST",
			path:   "/",
			request: handler.Input{
				URL: "",
			},
			id: "",
			want: want{
				status:      http.StatusBadRequest,
				contentType: "application/json",
				errorResponse: &handler.ErrorResponse{
					Error: "URL is required",
				},
			},
		},
		{
			name:   "service error",
			method: "POST",
			path:   "/",
			request: handler.Input{
				URL: "https://yandex.ru",
			},
			id: "",
			createFunc: func(url string, userID string) (string, error) {
				return "", errors.New("error")
			},
			want: want{
				status:      http.StatusInternalServerError,
				contentType: "application/json",
				errorResponse: &handler.ErrorResponse{
					Error: "Error while creating",
				},
			},
		},
		{
			name:   "already exists - conflict 409",
			method: "POST",
			path:   "/api/shorten",
			request: handler.Input{
				URL: "https://existing-url.com",
			},
			createFunc: func(url string, userID string) (string, error) {
				return "", &service.URLAlreadyExistsError{ID: "abc123"}
			},
			want: want{
				status:      http.StatusConflict,
				contentType: "application/json",
				response: &handler.Response{
					Result: config.Cfg.BaseURL + "/abc123",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc := &mocks.MockLinkService{
				CreateFunc: func(url string, userID string) (string, error) {
					if test.createFunc != nil {
						return test.createFunc(url, userID)
					}

					return test.id, nil
				},
				GetByURLFunc: func(url string) (model.Link, error) {
					return model.Link{
						BaseEntity: model.BaseEntity{ID: "abc123"},
						URL:        test.request.URL,
					}, nil
				},
			}

			jsonRequest, err := json.Marshal(test.request)
			require.NoError(t, err)

			request := httptest.NewRequest(test.method, test.path, bytes.NewReader(jsonRequest))
			request = request.WithContext(appctx.WithUserID(request.Context(), testUserID))

			w := httptest.NewRecorder()
			h := http.HandlerFunc(handler.GenerateJSON(svc, zap.NewNop(), config.Cfg, audit.NewBroadcaster(zap.NewNop())))

			h(w, request)

			result := w.Result()
			defer result.Body.Close()
			data, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			assert.Equal(t, test.want.status, result.StatusCode)
			assert.Contains(t, result.Header.Get("Content-Type"), test.want.contentType)

			if test.want.response != nil {
				var response handler.Response
				err = json.Unmarshal(data, &response)
				require.NoError(t, err)

				assert.Equal(t, *test.want.response, response)
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
