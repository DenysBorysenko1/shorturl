// Package middleware provides HTTP middleware components
// including authentication (JWT cookies) and response compression.
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"shorturl/internal/config"
	appctx "shorturl/internal/context"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func WithAuthCookie(logger *zap.Logger, config config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string

			cookie, err := r.Cookie("token")
			if err != nil {
				if errors.Is(err, http.ErrNoCookie) {
					var buildErr error
					tokenString, buildErr = BuildJWTString(logger)
					if buildErr != nil {
						logger.Error("Failed to build jwt token", zap.Error(buildErr))
						http.Error(w, "Internal server error", http.StatusInternalServerError)
						return
					}
					logger.Info("Set new auth cookie")
					SetAuthCookie(w, tokenString)
				} else {
					logger.Warn("Failed to read auth cookie", zap.Error(err))
					http.Error(w, "Bad request", http.StatusBadRequest)
					return
				}
			} else {
				logger.Info("Using existed auth cookie")
				tokenString = cookie.Value
			}

			userID, err := GetUserID(tokenString)
			if err != nil {
				logger.Warn("failed to parse user id from token", zap.Error(err))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			r = r.WithContext(appctx.WithUserID(r.Context(), userID))
			next.ServeHTTP(w, r)
		})
	}
}

func SetAuthCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:  "token",
		Value: token,
	}
	http.SetCookie(w, cookie)
}

func BuildJWTString(logger *zap.Logger) (string, error) {
	userID, err := generateRandomUserID()
	if err != nil {
		return "", err
	}

	expiration := time.Duration(config.Cfg.JWTExpiration) * time.Second
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(config.Cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(config.Cfg.JWTSecret), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("token invalid")
	}

	return claims.UserID, nil
}

func generateRandomUserID() (string, error) {
	b := make([]byte, 10)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
