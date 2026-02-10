package handler_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"shorturl/internal/handler"
	"shorturl/internal/service/mocks"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {

	type want struct {
		status      int
		contentType string
		response    string
	}

	tests := []struct {
		name       string
		id         string
		method     string
		path       string
		urlParam   string
		createFunc func(url string) (string, error)
		want       want
	}{
		{
			name:     "positive",
			method:   "POST",
			path:     "/",
			id:       "123",
			urlParam: "https://yandex.ru",
			want: want{
				status:      http.StatusCreated,
				contentType: "text/plain",
				response:    "http://localhost:8080/123",
			},
		},
		{
			name:     "incorrect method",
			method:   "GET",
			path:     "/",
			urlParam: "",
			id:       "",
			want: want{
				status:      http.StatusMethodNotAllowed,
				contentType: "text/plain",
				response:    "Only POST\n",
			},
		},
		{
			name:     "not pass param",
			method:   "POST",
			path:     "/",
			urlParam: "",
			id:       "",
			want: want{
				status:      http.StatusBadRequest,
				contentType: "text/plain",
				response:    "URL is required\n",
			},
		},
		{
			name:     "service error",
			method:   "POST",
			path:     "/",
			urlParam: "https://yandex.ru",
			id:       "",
			createFunc: func(url string) (string, error) {
				return "", errors.New("error")
			},
			want: want{
				status:      http.StatusInternalServerError,
				contentType: "text/plain",
				response:    "Error while creating\n",
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

			body := strings.NewReader(test.urlParam)
			request := httptest.NewRequest(test.method, test.path, body)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(handler.Generate(service, "http://localhost:8080"))

			h(w, request)

			result := w.Result()
			defer result.Body.Close()
			data, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			assert.Equal(t, test.want.status, result.StatusCode)
			assert.Contains(t, result.Header.Get("Content-Type"), test.want.contentType)
			assert.Equal(t, test.want.response, string(data))

		})
	}
}
