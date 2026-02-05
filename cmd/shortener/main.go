package main

import (
	"net/http"
	"shorturl/internal/config"
	"shorturl/internal/handler"
)

func main() {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/", handler.Generate)
	mux.HandleFunc("/{id}", handler.Retrieve)

	if err := http.ListenAndServe(string(config.AppPort), mux); err != nil {
		panic(err)
	}
}
