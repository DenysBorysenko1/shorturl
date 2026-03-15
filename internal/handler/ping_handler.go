package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

func Ping(log *zap.Logger, databasePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("sqlite", databasePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		err = db.PingContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Info("Database connection OK!")
		w.WriteHeader(http.StatusOK)
	}
}
