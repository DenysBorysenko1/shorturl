package main

import (
	"net/http"
	"github.com/DenysBorysenko1/shorturl/internal/handler"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handler.Generate)
	mux.HandleFunc(`/{id}/`, handler.Retrieve)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
