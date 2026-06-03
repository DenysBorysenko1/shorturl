package main

import (
	"database/sql"
	"net/http"
	"shorturl/internal/audit"
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

	broadcaster := initializeAuditBroadcaster(*logger.Log)
	router := newRouter(linkService, config.Cfg, broadcaster)

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

func initializeAuditBroadcaster(logger zap.Logger) *audit.Broadcaster {
	broadcaster := audit.NewBroadcaster(&logger)

	if config.Cfg.AuditFile != "" {
		fileObserver, err := audit.NewFileObserver(config.Cfg.AuditFile)
		if err != nil {
			logger.Error("Failed to initialize file audit observer", zap.Error(err))
		} else {
			broadcaster.Attach(fileObserver)
			logger.Info("File audit observer initialized", zap.String("path", config.Cfg.AuditFile))
		}
	}

	if config.Cfg.AuditURL != "" {
		httpObserver, err := audit.NewConcurrentHTTPObserver(config.Cfg.AuditURL)
		if err != nil {
			logger.Error("Failed to initialize HTTP audit observer", zap.Error(err))
		} else {
			broadcaster.Attach(httpObserver)
			logger.Info("HTTP audit observer initialized", zap.String("url", config.Cfg.AuditURL))
		}
	}

	return broadcaster
}

func newRouter(linkService service.LinkServiceInterface, cfg config.Config, broadcaster *audit.Broadcaster) chi.Router {
	router := chi.NewRouter()

	router.Use(logger.WithLogging)
	router.Use(middleware.WithAuthCookie(logger.Log, cfg))
	router.Use(middleware.WithCompress)

	router.Get("/{id}", handler.Retrieve(linkService, broadcaster))
	router.Get("/ping", handler.Ping(logger.Log, cfg))

	router.Post("/", handler.Generate(linkService, logger.Log, cfg, broadcaster))
	router.Post("/api/shorten", handler.GenerateJSON(linkService, logger.Log, cfg, broadcaster))
	router.Post("/api/shorten/batch", handler.GenerateBatch(linkService, logger.Log, cfg))

	router.Get("/api/user/urls", handler.ListUrls(linkService, logger.Log, cfg))

	router.Delete("/api/user/urls", handler.RemoveListUrls(linkService, logger.Log, cfg))

	return router
}
