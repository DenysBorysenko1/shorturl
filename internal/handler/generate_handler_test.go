package handler_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"shorturl/internal/config"
	appctx "shorturl/internal/context"
	"shorturl/internal/handler"
	"shorturl/internal/model"
	"shorturl/internal/service"
	"shorturl/internal/service/mocks"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const testUserID = "test-user-id"

func TestGenerate(t *testing.T) {
	config.Load()

	type want struct {
		status      int
		contentType string
		response    string
	}

	tests := []struct {
		name       string
		method     string
		path       string
		id         string
		urlParam   string
		createFunc func(url string, userId string) (string, error)
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
				response:    config.Cfg.BaseURL + "/123",
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
			createFunc: func(url string, userId string) (string, error) {
				assert.Equal(t, testUserID, userId)
				return "", errors.New("error")
			},
			want: want{
				status:      http.StatusInternalServerError,
				contentType: "text/plain",
				response:    "Error while creating\n",
			},
		},
		{
			name:     "unique violation conflict",
			method:   "POST",
			path:     "/",
			urlParam: "https://existing-url.com",
			createFunc: func(url string, userId string) (string, error) {
				assert.Equal(t, testUserID, userId)
				return "", &service.URLAlreadyExistsError{ID: "existingID"}
			},
			want: want{
				status:      http.StatusConflict,
				contentType: "text/plain",
				response:    config.Cfg.BaseURL + "/existingID",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc := &mocks.MockLinkService{
				CreateFunc: func(url string, userId string) (string, error) {
					if test.createFunc != nil {
						return test.createFunc(url, userId)
					}

					assert.Equal(t, testUserID, userId)
					return test.id, nil
				},
				GetByURLFunc: func(url string) (model.Link, error) {
					return model.Link{
						BaseEntity: model.BaseEntity{ID: "existingID"},
						URL:        url,
					}, nil
				},
			}

			body := strings.NewReader(test.urlParam)
			request := httptest.NewRequest(test.method, test.path, body)
			request = request.WithContext(appctx.WithUserID(request.Context(), testUserID))

			w := httptest.NewRecorder()
			h := http.HandlerFunc(handler.Generate(svc, zap.NewNop(), config.Cfg))

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
