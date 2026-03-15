package main

import (
	"net/http"
	"shorturl/internal/config"
	"shorturl/internal/handler"
	"shorturl/internal/logger"
	"shorturl/internal/middleware"
	"shorturl/internal/model"
	"shorturl/internal/repository"
	"shorturl/internal/service"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	config.Load()
	logger.Initialize(config.Cfg.LogLevel)

	router := chi.NewRouter()
	router.Use(logger.WithLogging)
	router.Use(middleware.WithCompress)

	fileRepository, err := repository.NewFileRepository[model.Link](config.Cfg.FileStorageURL)
	if err != nil {
		panic(err)
	}
	linkService := service.NewLinkService(fileRepository)

	router.Post("/", handler.Generate(linkService, config.Cfg.BaseURL))
	router.Post("/api/shorten", handler.GenerateJSON(linkService, config.Cfg.BaseURL))
	router.Get("/{id}", handler.Retrieve(linkService))
	router.Get("/ping", handler.Ping(logger.Log, config.Cfg.DatabaseDSN))

	logger.Log.Info("Starting server at", zap.String("address", config.Cfg.ServerAddress))
	if err := http.ListenAndServe(config.Cfg.ServerAddress, router); err != nil {
		panic(err)
	}
}
