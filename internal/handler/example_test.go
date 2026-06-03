package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"shorturl/internal/audit"
	"shorturl/internal/config"
	appctx "shorturl/internal/context"
	"shorturl/internal/dto"
	"shorturl/internal/handler"
	"shorturl/internal/model"
	"shorturl/internal/repository"
	"shorturl/internal/service"
	"strings"

	"go.uber.org/zap"
)

func ExampleGenerate() {
	config.Load()
	repo := repository.NewInMemoryRepository[model.Link]()
	logger, _ := zap.NewProduction()
	svc := service.NewLinkService(repo, *logger)
	h := handler.Generate(svc, logger, config.Cfg, audit.NewBroadcaster(logger))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	req = req.WithContext(appctx.WithUserID(req.Context(), "test-user"))
	w := httptest.NewRecorder()
	h(w, req)

	fmt.Println(w.Code)
	// Output:
	// 201
}

func ExampleGenerateJSON() {
	config.Load()
	repo := repository.NewInMemoryRepository[model.Link]()
	logger, _ := zap.NewProduction()
	svc := service.NewLinkService(repo, *logger)
	h := handler.GenerateJSON(svc, logger, config.Cfg, audit.NewBroadcaster(logger))

	body, _ := json.Marshal(handler.Input{URL: "https://example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	req = req.WithContext(appctx.WithUserID(req.Context(), "test-user"))
	w := httptest.NewRecorder()
	h(w, req)

	fmt.Println(w.Code)
	// Output:
	// 201
}

func ExampleGenerateBatch() {
	config.Load()
	repo := repository.NewInMemoryRepository[model.Link]()
	logger, _ := zap.NewProduction()
	svc := service.NewLinkService(repo, *logger)
	h := handler.GenerateBatch(svc, logger, config.Cfg)

	batch := []dto.RequestBatchItem{
		{CorrelationID: "id1", OriginalURL: "https://example1.com"},
		{CorrelationID: "id2", OriginalURL: "https://example2.com"},
	}
	body, _ := json.Marshal(batch)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
	req = req.WithContext(appctx.WithUserID(req.Context(), "test-user"))
	w := httptest.NewRecorder()
	h(w, req)

	fmt.Println(w.Code)
	// Output:
	// 201
}

func ExampleRetrieve() {
	repo := repository.NewInMemoryRepository[model.Link]()
	logger, _ := zap.NewProduction()
	svc := service.NewLinkService(repo, *logger)

	_ = repo.Create(model.Link{
		BaseEntity: model.BaseEntity{ID: "abc123"},
		URL:        "https://example.com",
	})
	id := "abc123"

	h := handler.Retrieve(svc, audit.NewBroadcaster(logger))
	req := httptest.NewRequest(http.MethodGet, "/"+id, nil)
	w := httptest.NewRecorder()
	h(w, req)

	fmt.Println(w.Code)
	// Output:
	// 404
}

func ExampleListUrls() {
	config.Load()
	repo := repository.NewInMemoryRepository[model.Link]()
	logger, _ := zap.NewProduction()
	svc := service.NewLinkService(repo, *logger)

	svc.Create("https://example1.com", "test-user")
	svc.Create("https://example2.com", "test-user")

	h := handler.ListUrls(svc, logger, config.Cfg)
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req = req.WithContext(appctx.WithUserID(req.Context(), "test-user"))
	w := httptest.NewRecorder()
	h(w, req)

	fmt.Println(w.Code)
	fmt.Println(strings.Contains(w.Body.String(), "example1.com"))
	fmt.Println(strings.Contains(w.Body.String(), "example2.com"))
	// Output:
	// 200
	// true
	// true
}

func ExampleRemoveListUrls() {
	repo := repository.NewInMemoryRepository[model.Link]()
	logger, _ := zap.NewProduction()
	svc := service.NewLinkService(repo, *logger)

	id1, _ := svc.Create("https://example1.com", "test-user")
	id2, _ := svc.Create("https://example2.com", "test-user")

	h := handler.RemoveListUrls(svc, logger, config.Cfg)
	ids := []string{id1, id2}
	body, _ := json.Marshal(ids)
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
	req = req.WithContext(appctx.WithUserID(req.Context(), "test-user"))
	w := httptest.NewRecorder()
	h(w, req)

	fmt.Println(w.Code)
	// Output:
	// 202
}

func ExamplePing() {
	config.Load()
	logger, _ := zap.NewProduction()
	h := handler.Ping(logger, config.Cfg)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	h(w, req)

	fmt.Println(w.Code)
	// Output:
	// 500
}
