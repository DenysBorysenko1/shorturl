package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"shorturl/internal/handler"
	"shorturl/internal/service/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateJSON(t *testing.T) {

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
		createFunc func(url string) (string, error)
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
					Result: "http://localhost:8080/123",
				},
			},
		},
		{
			name:   "incorrect method",
			method: "GET",
			path:   "/",
			id:     "",
			request: handler.Input{
				URL: "https://yandex.ru",
			},
			want: want{
				status:      http.StatusMethodNotAllowed,
				contentType: "application/json",
				errorResponse: &handler.ErrorResponse{
					Error: "Only POST",
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
			createFunc: func(url string) (string, error) {
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &mocks.MockLinkService{
				CreateFunc: func(url string) (string, error) {
					if test.createFunc != nil {
						return test.createFunc(url)
					}

					return test.id, nil
				},
			}

			jsonRequest, err := json.Marshal(test.request)
			require.NoError(t, err)

			request := httptest.NewRequest(test.method, test.path, bytes.NewReader(jsonRequest))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(handler.GenerateJSON(service, "http://localhost:8080"))

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
