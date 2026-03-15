package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"go.uber.org/zap"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Ping(log *zap.Logger, databaseDSN string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("pgx", databaseDSN)
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
