package handler_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"shorturl/internal/config"
	appctx "shorturl/internal/context"
	"shorturl/internal/dto"
	"shorturl/internal/handler"
	"shorturl/internal/middleware"
	"shorturl/internal/model"
	"shorturl/internal/service/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestListUrls(t *testing.T) {
	config.Load()

	type want struct {
		status          int
		contentType     string
		emptyBody       bool
		errorSubstring  string
		setCookiePrefix string
		response        []dto.ResponseListItem
	}

	tests := []struct {
		name             string
		userID           string
		useMiddleware    bool
		getAllByUserIDFn func(userID string) ([]model.Link, error)
		want             want
	}{
		{
			name: "no user in context",
			want: want{
				status: http.StatusInternalServerError,
			},
		},
		{
			name:   "user in context empty list",
			userID: "test-user-id",
			getAllByUserIDFn: func(userID string) ([]model.Link, error) {
				assert.Equal(t, "test-user-id", userID)
				return nil, nil
			},
			want: want{
				status:    http.StatusNoContent,
				emptyBody: true,
			},
		},
		{
			name:   "user with links returns json",
			userID: "test-user-id",
			getAllByUserIDFn: func(userID string) ([]model.Link, error) {
				assert.Equal(t, "test-user-id", userID)
				return []model.Link{
					{BaseEntity: model.BaseEntity{ID: "abc123"}, URL: "https://one.com"},
					{BaseEntity: model.BaseEntity{ID: "xyz9"}, URL: "https://two.com"},
				}, nil
			},
			want: want{
				status:      http.StatusOK,
				contentType: "application/json",
				response: []dto.ResponseListItem{
					{ShortURL: config.Cfg.BaseURL + "/abc123", OriginalURL: "https://one.com"},
					{ShortURL: config.Cfg.BaseURL + "/xyz9", OriginalURL: "https://two.com"},
				},
			},
		},
		{
			name:   "service error",
			userID: "test-user-id",
			getAllByUserIDFn: func(userID string) ([]model.Link, error) {
				return nil, errors.New("db error")
			},
			want: want{
				status:         http.StatusBadRequest,
				contentType:    "application/json",
				errorSubstring: "db error",
			},
		},
		{
			name:          "middleware without cookie issues token and reaches handler",
			useMiddleware: true,
			getAllByUserIDFn: func(userID string) ([]model.Link, error) {
				assert.NotEmpty(t, userID)
				return nil, nil
			},
			want: want{
				status:          http.StatusNoContent,
				emptyBody:       true,
				setCookiePrefix: "token=",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

			svc := &mocks.MockLinkService{}
			if tt.getAllByUserIDFn != nil {
				svc.GetAllByUserIDFunc = tt.getAllByUserIDFn
			}

			if tt.userID != "" && !tt.useMiddleware {
				req = req.WithContext(appctx.WithUserID(req.Context(), tt.userID))
			}

			w := httptest.NewRecorder()

			var h http.Handler = http.HandlerFunc(handler.ListUrls(svc, logger, config.Cfg))
			if tt.useMiddleware {
				h = middleware.WithAuthCookie(logger, config.Cfg)(h)
			}

			h.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.status, res.StatusCode)

			if tt.want.contentType != "" {
				assert.Contains(t, res.Header.Get("Content-Type"), tt.want.contentType)
			}

			if tt.want.setCookiePrefix != "" {
				assert.Contains(t, res.Header.Get("Set-Cookie"), tt.want.setCookiePrefix)
			}

			needBody := tt.want.emptyBody || tt.want.errorSubstring != "" || tt.want.response != nil
			if !needBody {
				return
			}

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if tt.want.emptyBody {
				assert.Empty(t, body)
			}

			if tt.want.errorSubstring != "" {
				var er handler.ErrorResponse
				require.NoError(t, json.Unmarshal(body, &er))
				assert.Contains(t, er.Error, tt.want.errorSubstring)
			}

			if tt.want.response != nil {
				var got []dto.ResponseListItem
				require.NoError(t, json.Unmarshal(body, &got))
				assert.Equal(t, tt.want.response, got)
			}
		})
	}
}
