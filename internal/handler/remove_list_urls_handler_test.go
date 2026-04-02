package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"shorturl/internal/config"
	appctx "shorturl/internal/context"
	"shorturl/internal/handler"
	"shorturl/internal/service/mocks"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRemoveListUrls(t *testing.T) {
	config.Load()

	userID := "test-user-id"

	tests := []struct {
		name              string
		userInContext     bool
		body              string
		enqueueErr        error
		wantEnqueueCalled bool
		wantIDs           []string
		wantUserID        string
		wantStatus        int
		wantBodySubstring string
		wantErrorResponse bool
	}{
		{
			name:              "no user in context",
			userInContext:     false,
			body:              `["id1","id2"]`,
			wantStatus:        http.StatusInternalServerError,
			wantBodySubstring: "Internal server error",
			wantEnqueueCalled: false,
		},
		{
			name:              "bad json",
			userInContext:     true,
			body:              `not-json`,
			wantStatus:        http.StatusBadRequest,
			wantErrorResponse: true,
			wantEnqueueCalled: false,
		},
		{
			name:              "positive",
			userInContext:     true,
			body:              `["id1","id2"]`,
			wantStatus:        http.StatusAccepted,
			wantEnqueueCalled: true,
			wantIDs:           []string{"id1", "id2"},
			wantUserID:        userID,
		},
		{
			name:              "service error ignored",
			userInContext:     true,
			body:              `["id1","id2"]`,
			enqueueErr:        errors.New("db error"),
			wantStatus:        http.StatusAccepted,
			wantEnqueueCalled: true,
			wantIDs:           []string{"id1", "id2"},
			wantUserID:        userID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			svc := &mocks.MockLinkService{}

			var reqBody io.Reader = strings.NewReader(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/user/urls/delete", reqBody)
			if tt.userInContext {
				req = req.WithContext(appctx.WithUserID(req.Context(), userID))
			}

			w := httptest.NewRecorder()

			called := false
			var gotIDs []string
			var gotUserID string
			svc.EnqueueDeleteFunc = func(ids []string, uid string) error {
				called = true
				gotIDs = ids
				gotUserID = uid
				return tt.enqueueErr
			}

			h := handler.RemoveListUrls(svc, logger, config.Cfg)
			h(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)

			data, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if !tt.wantEnqueueCalled {
				assert.False(t, called, "EnqueueDelete should not be called")
			} else {
				assert.True(t, called, "EnqueueDelete should be called")
				assert.Equal(t, tt.wantIDs, gotIDs)
				assert.Equal(t, tt.wantUserID, gotUserID)
			}

			if tt.wantErrorResponse {
				var er handler.ErrorResponse
				require.NoError(t, json.Unmarshal(data, &er))
				assert.NotEmpty(t, er.Error)
				assert.Contains(t, res.Header.Get("Content-Type"), "application/json")
				return
			}

			if tt.wantBodySubstring != "" {
				assert.Contains(t, string(data), tt.wantBodySubstring)
				return
			}

			assert.Empty(t, bytes.TrimSpace(data))
		})
	}
}
