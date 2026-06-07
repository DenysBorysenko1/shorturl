package pool

import (
	"sync"
	"testing"

	"go.uber.org/mock/gomock"
)

type countResetter struct {
	count int
	data  []string
}

func (c *countResetter) Reset() {
	c.count++
	c.data = nil
}

type sliceResetter struct {
	data []string
}

func (s *sliceResetter) Reset() {
	s.data = nil
}

func TestPool_New(t *testing.T) {
	factory := func() *countResetter {
		return &countResetter{}
	}

	p := New(factory)

	if p == nil {
		t.Fatal("Pool is nil")
	}
}

func TestPool_Get(t *testing.T) {
	factory := func() *countResetter {
		return &countResetter{data: []string{"a", "b"}}
	}

	p := New(factory)

	item := p.Get()

	if item == nil {
		t.Fatal("Get returned nil")
	}

	if len(item.data) != 2 {
		t.Errorf("Expected len 2, got %d", len(item.data))
	}
}

func TestPool_Put(t *testing.T) {
	factory := func() *countResetter {
		return &countResetter{data: []string{"a", "b"}}
	}

	p := New(factory)

	item := p.Get()
	item.data = append(item.data, "c")

	p.Put(item)

	if item.data != nil {
		t.Errorf("Expected data to be nil after Reset, got %v", item.data)
	}
	if item.count != 1 {
		t.Errorf("Expected count to be 1 after Reset, got %d", item.count)
	}
}

func TestPool_Concurrent(t *testing.T) {
	factory := func() *countResetter {
		return &countResetter{data: []string{"a", "b"}}
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
				item.data = []string{"test"}
				p.Put(item)
			}
		}()
	}

	wg.Wait()

	t.Log("Concurrent access works correctly")
}

func TestPool_WithDifferentTypes(t *testing.T) {
	resettableFactory := func() *sliceResetter {
		return &sliceResetter{
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

func TestPool_NilFactory(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when factory is nil, but got none")
		} else if r != "pool: factory function cannot be nil" {
			t.Errorf("Expected panic message 'pool: factory function cannot be nil', got '%v'", r)
		}
	}()

	New[*countResetter](nil)
	t.Error("Test should have panicked")
}

func TestPool_WithMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockResetter := NewMockResetter(ctrl)
	mockResetter.EXPECT().Reset().Times(2)

	callCount := 0
	factory := func() Resetter {
		callCount++
		if callCount == 1 {
			return mockResetter
		}
		return mockResetter
	}

	p := New(factory)

	item := p.Get()
	p.Put(item)

	item2 := p.Get()
	p.Put(item2)
}
