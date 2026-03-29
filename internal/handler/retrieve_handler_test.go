package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"shorturl/internal/handler"
	"shorturl/internal/service/mocks"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestRetrieve(t *testing.T) {

	type want struct {
		status   int
		location string
		response string
	}

	tests := []struct {
		name       string
		id         string
		path       string
		method     string
		urlParam   string
		createFunc func(url string, userId string) (string, error)
		getFunc    func(id string) (string, error)
		want       want
	}{
		{
			name:   "positive",
			method: "GET",
			path:   "/123",
			getFunc: func(id string) (string, error) {
				return "https://yandex.ru", nil
			},
			want: want{
				status:   http.StatusTemporaryRedirect,
				location: "https://yandex.ru",
			},
		},
		{
			name:   "method incorrect",
			method: "POST",
			path:   "/123",
			want: want{
				status:   http.StatusMethodNotAllowed,
				location: "",
			},
		},
		{
			name:   "not passed id",
			method: "GET",
			path:   "/",
			want: want{
				status:   http.StatusNotFound,
				location: "",
			},
		},
		{
			name:   "error while processing",
			method: "GET",
			path:   "/123",
			getFunc: func(id string) (string, error) {
				return "", errors.New("error")
			},
			want: want{
				status:   http.StatusInternalServerError,
				location: "",
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

					return test.id, nil
				},
				GetFunc: func(id string) (string, error) {
					if test.getFunc != nil {
						return test.getFunc(id)
					}

					return test.want.location, nil
				},
			}
			r := chi.NewRouter()
			r.Get("/{id}", handler.Retrieve(svc))
			request := httptest.NewRequest(test.method, "http://localhost"+test.path, nil)
			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, request)

			result := recorder.Result()
			defer result.Body.Close()

			assert.Equal(t, test.want.status, result.StatusCode)
			assert.Contains(t, result.Header.Get("Location"), test.want.location)

		})
	}

}
