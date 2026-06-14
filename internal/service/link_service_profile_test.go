package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"shorturl/internal/model"
	"shorturl/internal/repository"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestProfileLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping profiling test in short mode")
	}

	core, _ := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	repo := repository.NewInMemoryRepository[model.Link]()
	svc := NewLinkService(repo, *logger)

	t.Log("Generating 10,000 short URLs...")
	ids := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		id, err := generateID()
		if err != nil {
			t.Errorf("Error generating ID: %v", err)
			continue
		}
		ids[i] = id

		link := model.Link{
			BaseEntity: model.BaseEntity{ID: id},
			URL:        fmt.Sprintf("https://example.com/page-%d", i),
			CreatedBy:  "user123",
		}
		if err := repo.Create(link); err != nil {
			t.Errorf("Error creating link: %v", err)
		}
	}

	t.Log("Performing 10,000 get operations...")
	for i := 0; i < 10000; i++ {
		_, err := svc.Get(ids[i])
		if err != nil {
			t.Errorf("Error getting link: %v", err)
		}
	}

	t.Log("Performing 1,000 GetAllByUserID operations...")
	for i := 0; i < 1000; i++ {
		_, err := svc.GetAllByUserID("user123")
		if err != nil {
			t.Errorf("Error getting all links: %v", err)
		}
	}

	t.Log("Performing 100 batch operations...")
	batch := make([]model.Link, 100)
	for i := 0; i < 100; i++ {
		id, err := generateID()
		if err != nil {
			t.Errorf("Error generating ID: %v", err)
			continue
		}
		batch[i] = model.Link{
			BaseEntity: model.BaseEntity{ID: id},
			URL:        fmt.Sprintf("https://example.com/batch-%d", i),
			CreatedBy:  "user456",
		}
	}
	for i := 0; i < 100; i++ {
		if err := svc.CreateMany(batch); err != nil {
			t.Errorf("Error creating batch: %v", err)
		}
	}

	t.Log("Performing delete operations...")
	deleteIDs := ids[:1000]
	for i := 0; i < 10; i++ {
		if err := svc.EnqueueDelete(deleteIDs, "user123"); err != nil {
			t.Errorf("Error enqueueing delete: %v", err)
		}
	}

	t.Log("Waiting for delete operations to complete...")
	for _, id := range deleteIDs[:10] {
		link, err := repo.GetByID(id)
		if err == nil {
			if !link.IsDeleted {
				t.Logf("Link %s is not yet deleted", id)
			}
		}
	}

	t.Log("Workload completed")
}

func generateID() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
