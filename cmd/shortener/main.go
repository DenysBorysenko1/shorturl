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
	r := chi.NewRouter()

	inMemoryRepository := repository.NewInMemoryRepository[model.Link]()
	linkService := service.NewLinkService(inMemoryRepository)

	r.Post("/", handler.Generate(linkService))
	r.Get("/{id}", handler.Retrieve(linkService))

	fmt.Println("Starting http://localhost:" + string(config.AppPort))
	if err := http.ListenAndServe(":"+string(config.AppPort), r); err != nil {
		panic(err)
	}
}
