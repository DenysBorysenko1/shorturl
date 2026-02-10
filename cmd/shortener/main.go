package main

import (
	"fmt"
	"net/http"
	"shorturl/internal/config"
	"shorturl/internal/handler"
	"shorturl/internal/model"
	"shorturl/internal/repository"
	"shorturl/internal/service"

	"github.com/go-chi/chi/v5"
)

func main() {
	config.Load()

	r := chi.NewRouter()

	inMemoryRepository := repository.NewInMemoryRepository[model.Link]()
	linkService := service.NewLinkService(inMemoryRepository)

	r.Post("/", handler.Generate(linkService, config.Cfg.BaseURL))
	r.Get("/{id}", handler.Retrieve(linkService))

	fmt.Println("Starting server at", config.Cfg.ServerAddress)
	if err := http.ListenAndServe(config.Cfg.ServerAddress, r); err != nil {
		panic(err)
	}
}
