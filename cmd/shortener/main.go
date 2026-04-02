package main

import (
	"database/sql"
	"net/http"
	"shorturl/internal/config"
	dbpkg "shorturl/internal/db"
	"shorturl/internal/handler"
	"shorturl/internal/logger"
	"shorturl/internal/middleware"
	"shorturl/internal/model"
	"shorturl/internal/repository"
	"shorturl/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5"
)

func main() {
	config.Load()
	logger.Initialize(config.Cfg.LogLevel)

	linkRepository := initializeLinkRepository(*logger.Log)
	linkService := service.NewLinkService(linkRepository, *logger.Log)

	router := newRouter(linkService, config.Cfg)

	logger.Log.Info("Starting server at", zap.String("address", config.Cfg.ServerAddress))
	if err := http.ListenAndServe(config.Cfg.ServerAddress, router); err != nil {
		logger.Log.Fatal("Server stopped with error", zap.Error(err))
	}
}

func initializeLinkRepository(logger zap.Logger) repository.Repository[model.Link] {
	if config.Cfg.DatabaseDSN != "" {
		db, err := sql.Open("pgx", config.Cfg.DatabaseDSN)
		if err != nil {
			logger.Fatal("Failed to open database", zap.Error(err))
		}
		if err = db.Ping(); err != nil {
			logger.Fatal("Failed to ping database", zap.Error(err))
		}

		if err := dbpkg.RunMigrations(db); err != nil {
			logger.Fatal("Failed to run migrations", zap.Error(err))
		}

		logger.Info("While initialize link repository POSTGRES source had chosen")

		sqlxDB := sqlx.NewDb(db, "pgx")
		return repository.NewPostgresRepository[model.Link](sqlxDB)
	}

	if config.Cfg.FileStorageURL != "" {
		repository, err := repository.NewFileRepository[model.Link](config.Cfg.FileStorageURL)
		if err != nil {
			logger.Error("Failed to init file repository: %v", zap.Error(err))
		}
		logger.Info("While initialize link repository FILE source had chosen")
		return repository
	}

	logger.Info("While initialize link repository IN MEMORY source had chosen")
	return repository.NewInMemoryRepository[model.Link]()
}

func newRouter(linkService service.LinkServiceInterface, config config.Config) chi.Router {
	router := chi.NewRouter()

	router.Use(logger.WithLogging)
	router.Use(middleware.WithAuthCookie(logger.Log, config))
	router.Use(middleware.WithCompress)

	router.Get("/{id}", handler.Retrieve(linkService))
	router.Get("/ping", handler.Ping(logger.Log, config))

	router.Post("/", handler.Generate(linkService, logger.Log, config))
	router.Post("/api/shorten", handler.GenerateJSON(linkService, logger.Log, config))
	router.Post("/api/shorten/batch", handler.GenerateBatch(linkService, logger.Log, config))

	router.Get("/api/user/urls", handler.ListUrls(linkService, logger.Log, config))

	router.Delete("/api/user/urls", handler.RemoveListUrls(linkService, logger.Log, config))

	return router
}
