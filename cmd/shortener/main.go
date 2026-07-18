package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"shorturl/internal/audit"
	"shorturl/internal/config"
	dbpkg "shorturl/internal/db"
	grpcserver "shorturl/internal/grpc"
	"shorturl/internal/handler"
	"shorturl/internal/logger"
	"shorturl/internal/middleware"
	"shorturl/internal/model"
	"shorturl/internal/repository"
	"shorturl/internal/service"
	pb "shorturl/proto"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	_ "github.com/jackc/pgx/v5"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	config.Load()
	logger.Initialize(config.Cfg.LogLevel)

	linkRepository, statsProvider := initializeLinkRepository(*logger.Log)
	linkService := service.NewLinkService(linkRepository, *logger.Log)

	broadcaster := initializeAuditBroadcaster(*logger.Log)
	router := newRouter(linkService, statsProvider, config.Cfg, broadcaster)

	logger.Log.Info("Starting HTTP server at", zap.String("address", config.Cfg.ServerAddress))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	var srv *http.Server
	if config.Cfg.EnableHTTPS {
		srv = createHTTPSServer(router)
	} else {
		srv = createHTTPServer(router)
	}

	go func() {
		var err error
		if config.Cfg.EnableHTTPS {
			err = srv.ListenAndServeTLS(config.Cfg.TLSCertFile, config.Cfg.TLSKeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatal("HTTP server stopped with error", zap.Error(err))
		}
	}()

	// gRPC server
	var grpcSrv *grpc.Server
	if config.Cfg.GRPCAddress != "" {
		grpcSrv = grpc.NewServer(grpc.UnaryInterceptor(grpcserver.AuthInterceptor()))
		grpcSvc := grpcserver.NewServer(linkService, config.Cfg, logger.Log)
		pb.RegisterShortenerServiceServer(grpcSrv, grpcSvc)

		go func() {
			grpcListener, netErr := net.Listen("tcp", config.Cfg.GRPCAddress)
			if netErr != nil {
				logger.Log.Fatal("Failed to listen gRPC", zap.Error(netErr))
			}
			logger.Log.Info("Starting gRPC server at", zap.String("address", config.Cfg.GRPCAddress))
			if err := grpcSrv.Serve(grpcListener); err != nil {
				logger.Log.Fatal("gRPC server stopped with error", zap.Error(err))
			}
		}()
	}

	<-ctx.Done()
	logger.Log.Info("Received signal, shutting down gracefully")

	const shutdownTimeout = 30 * time.Second
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if grpcSrv != nil {
		grpcSrv.GracefulStop()
		logger.Log.Info("gRPC server stopped")
	}

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("HTTP server shutdown error", zap.Error(err))
	}
	logger.Log.Info("HTTP server stopped")

	linkService.Close()
	logger.Log.Info("Link service closed")

	broadcaster.Close()
	logger.Log.Info("Audit broadcaster closed")

	if sqlDB, ok := linkRepository.(interface{ Close() error }); ok {
		if err := sqlDB.Close(); err != nil {
			logger.Log.Error("Database close error", zap.Error(err))
		} else {
			logger.Log.Info("Database connection closed")
		}
	}

	logger.Log.Info("Server shutdown complete")
}

func createHTTPServer(router chi.Router) *http.Server {
	return &http.Server{
		Addr:    config.Cfg.ServerAddress,
		Handler: router,
	}
}

func createHTTPSServer(router chi.Router) *http.Server {
	logger.Log.Info("HTTPS mode enabled",
		zap.String("cert_file", config.Cfg.TLSCertFile),
		zap.String("key_file", config.Cfg.TLSKeyFile),
	)

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	return &http.Server{
		Addr:      config.Cfg.ServerAddress,
		Handler:   router,
		TLSConfig: tlsConfig,
	}
}

func initializeLinkRepository(logger zap.Logger) (repository.Repository[model.Link], repository.StatsProvider) {
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
		repo := repository.NewPostgresRepository[model.Link](sqlxDB)
		return repo, repo
	}

	if config.Cfg.FileStorageURL != "" {
		repository, err := repository.NewFileRepository[model.Link](config.Cfg.FileStorageURL)
		if err != nil {
			logger.Error("Failed to init file repository: %v", zap.Error(err))
		}
		logger.Info("While initialize link repository FILE source had chosen")
		return repository, repository
	}

	logger.Info("While initialize link repository IN MEMORY source had chosen")
	repo := repository.NewInMemoryRepository[model.Link]()
	return repo, repo
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

func newRouter(linkService service.LinkServiceInterface, statsProvider repository.StatsProvider, cfg config.Config, broadcaster *audit.Broadcaster) chi.Router {
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

	router.Get("/api/internal/stats", handler.InternalStats(statsProvider, cfg))

	return router
}
