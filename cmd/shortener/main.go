package main

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var (
	links  = make(map[string]string) // id → куда редиректить
	lastID int
	mu     sync.Mutex
)

const baseURL = "http://localhost:8080"

func generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST", http.StatusMethodNotAllowed)
		return
	}
	body, _ := io.ReadAll(io.LimitReader(r.Body, 2048))
	_ = r.Body.Close()
	url := strings.TrimSpace(string(body))
	if url == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	lastID++
	id := strconv.Itoa(lastID)
	links[id] = url
	mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(baseURL + "/" + id))
}

func retrieveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET", http.StatusMethodNotAllowed)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	mu.Lock()
	url, ok := links[id]
	mu.Unlock()

	if !ok {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", generateHandler)
	mux.HandleFunc("/{id}", retrieveHandler)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
