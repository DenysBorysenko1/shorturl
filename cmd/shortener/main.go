package main

import (
	"fmt"
	"net/http"
	"shorturl/internal/config"
	"shorturl/internal/handler"
	"shorturl/internal/logger"
	"shorturl/internal/model"
	"shorturl/internal/repository"
	"shorturl/internal/service"

	"github.com/go-chi/chi/v5"
)

func main() {
	config.Load()
	logger.Initialize(config.Cfg.LogLevel)

	router := chi.NewRouter()
	router.Use(logger.WithLogging)

	inMemoryRepository := repository.NewInMemoryRepository[model.Link]()
	linkService := service.NewLinkService(inMemoryRepository)

	router.Post("/", handler.Generate(linkService, config.Cfg.BaseURL))
	router.Get("/{id}", handler.Retrieve(linkService))

	fmt.Println("Starting server at", config.Cfg.ServerAddress)
	if err := http.ListenAndServe(config.Cfg.ServerAddress, router); err != nil {
		panic(err)
	}
}
