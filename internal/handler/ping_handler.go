package handler

import (
	"context"
	"database/sql"
	"net/http"
	"shorturl/internal/config"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// Ping creates and returns an HTTP handler for checking database connectivity.
//
// The handler accepts GET requests and attempts to connect to the database configured
// in the application config. It returns 200 OK if the connection is successful, or
// 500 Internal Server Error if the connection fails or times out. A 1-second timeout
// is applied to the database ping operation.
//
// Request format:
//
//	GET /ping
//
// Response formats:
//
//	Success (200 OK): Database connection is healthy
//	Error (500 Internal Server Error): Database connection failed or timed out
//
// Parameters:
//   - log: Logger for logging connection status
//   - config: Configuration containing DatabaseDSN for database connection
//
// Returns:
//   - http.HandlerFunc: Handler function for database health check
func Ping(log *zap.Logger, config config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("pgx", config.DatabaseDSN)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		log.Info("Database connection OK!")
		w.WriteHeader(http.StatusOK)
	}
}
