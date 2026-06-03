package service

import (
	"shorturl/internal/model"
	"shorturl/internal/repository"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func BenchmarkLinkService_Create(b *testing.B) {
	core, _ := observer.New(zap.InfoLevel)
	logger := *zap.New(core)

	repo := repository.NewInMemoryRepository[model.Link]()
	svc := NewLinkService(repo, logger)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, err := svc.Create("https://example.com/test-url", "user123")
		if err != nil {
			b.Fatalf("Create failed: %v", err)
		}
	}
}

func BenchmarkLinkService_Get(b *testing.B) {
	core, _ := observer.New(zap.InfoLevel)
	logger := *zap.New(core)

	repo := repository.NewInMemoryRepository[model.Link]()
	svc := NewLinkService(repo, logger)

	id, err := svc.Create("https://example.com/test-url", "user123")
	if err != nil {
		b.Fatalf("Create failed: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, err := svc.Get(id)
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}

func BenchmarkLinkService_CreateMany(b *testing.B) {
	core, _ := observer.New(zap.InfoLevel)
	logger := *zap.New(core)

	repo := repository.NewInMemoryRepository[model.Link]()
	svc := NewLinkService(repo, logger)

	batchSize := 100
	links := make([]model.Link, batchSize)
	for i := 0; i < batchSize; i++ {
		links[i] = model.Link{
			BaseEntity: model.BaseEntity{ID: string(rune('a' + i%26))},
			URL:        "https://example.com/test-url",
			CreatedBy:  "user123",
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		err := svc.CreateMany(links)
		if err != nil {
			b.Fatalf("CreateMany failed: %v", err)
		}
	}
}

func BenchmarkLinkService_GetAllByUserID(b *testing.B) {
	core, _ := observer.New(zap.InfoLevel)
	logger := *zap.New(core)

	repo := repository.NewInMemoryRepository[model.Link]()
	svc := NewLinkService(repo, logger)

	for i := 0; i < 100; i++ {
		_, err := svc.Create("https://example.com/test-url", "user123")
		if err != nil {
			b.Fatalf("Create failed: %v", err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, err := svc.GetAllByUserID("user123")
		if err != nil {
			b.Fatalf("GetAllByUserID failed: %v", err)
		}
	}
}
