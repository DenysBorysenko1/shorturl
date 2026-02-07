package handler_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"shorturl/internal/config"
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
		path       string
		urlParam   string
		createFunc func(url string) (string, error)
		want       want
	}{
		{
			name:     "positive",
			path:     "/",
			id:       "123",
			urlParam: "https://yandex.ru",
			want: want{
				status:      http.StatusCreated,
				contentType: "text/plain",
				response:    config.AppBaseURL + "/123",
			},
		},
		{
			name:     "not pass param",
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
			request := httptest.NewRequest(http.MethodPost, test.path, body)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(handler.Generate(service))

			h(w, request)

			result := w.Result()
			data, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			assert.Equal(t, test.want.status, result.StatusCode)
			assert.Contains(t, result.Header.Get("Content-Type"), test.want.contentType)
			assert.Equal(t, test.want.response, string(data))

		})
	}
}
