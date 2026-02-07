package main

import (
	"fmt"
	"net/http"
	"shorturl/internal/config"
	"shorturl/internal/handler"
	"shorturl/internal/model"
	"shorturl/internal/repository"
	"shorturl/internal/service"
)

func main() {
	mux := http.NewServeMux()

	inMemoryRepository := repository.NewInMemoryRepository[model.Link]()
	linkService := service.NewLinkService(inMemoryRepository)

	mux.HandleFunc("/", handler.Generate(linkService))
	mux.HandleFunc("/{id}", handler.Retrieve(linkService))

	fmt.Println("Starting http://localhost:" + string(config.AppPort))
	if err := http.ListenAndServe(":"+string(config.AppPort), mux); err != nil {
		panic(err)
	}
}
