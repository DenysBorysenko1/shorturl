package handler

import (
	"fmt"
	"io"
	"net/http"
	"shorturl/internal/config"
	"strconv"
	"strings"
)


func Generate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 2048))
	if err != nil {
		fmt.Println(err)
	}

	defer r.Body.Close()
	url := strings.TrimSpace(string(body))
	if url == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	//todo: service.save()

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(config.AppBaseUrl + "/" + id))
}