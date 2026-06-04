package pool

import (
	"sync"
	"testing"
)

type MockResetter struct {
	Value int
}

func (m *MockResetter) Reset() {
	m.Value = 0
}

type ResettableSlice struct {
	data []string
}

func (r *ResettableSlice) Reset() {
	r.data = nil
}

func TestPool_New(t *testing.T) {
	factory := func() *MockResetter {
		return &MockResetter{Value: 42}
	}

	p := New(factory)

	if p == nil {
		t.Fatal("Pool is nil")
	}
}

func TestPool_Get(t *testing.T) {
	factory := func() *MockResetter {
		return &MockResetter{Value: 42}
	}

	p := New(factory)

	item := p.Get()

	if item == nil {
		t.Fatal("Get returned nil")
	}

	if item.Value != 42 {
		t.Errorf("Expected Value 42, got %d", item.Value)
	}
}

func TestPool_Put(t *testing.T) {
	factory := func() *MockResetter {
		return &MockResetter{Value: 0}
	}

	p := New(factory)

	item := p.Get()
	item.Value = 100

	p.Put(item)

	if item.Value != 0 {
		t.Errorf("Expected Value 0 after Reset, got %d", item.Value)
	}

	item2 := p.Get()
	if item == item2 {
		t.Log("Pool reuses objects correctly")
	}
}

func TestPool_Concurrent(t *testing.T) {
	factory := func() *MockResetter {
		return &MockResetter{Value: 42}
	}

	p := New(factory)

	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				item := p.Get()
				item.Value = j
				p.Put(item)
			}
		}()
	}

	wg.Wait()

	t.Log("Concurrent access works correctly")
}

func TestPool_WithDifferentTypes(t *testing.T) {
	resettableFactory := func() *ResettableSlice {
		return &ResettableSlice{
			data: []string{"a", "b", "c"},
		}
	}

	pool2 := New(resettableFactory)

	item := pool2.Get()
	if len(item.data) != 3 {
		t.Errorf("Expected len 3, got %d", len(item.data))
	}

	pool2.Put(item)
	if len(item.data) != 0 {
		t.Errorf("Expected len 0 after Reset, got %d", len(item.data))
	}
}
