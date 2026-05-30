package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type HTTPObserver struct {
	url    string
	client *http.Client
}

func NewHTTPObserver(url string) (*HTTPObserver, error) {
	if url == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	return &HTTPObserver{
		url: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

func (h *HTTPObserver) LogEvent(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, h.url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

type ConcurrentHTTPObserver struct {
	observer *HTTPObserver
	wg       *sync.WaitGroup
}

func NewConcurrentHTTPObserver(url string) (*ConcurrentHTTPObserver, error) {
	observer, err := NewHTTPObserver(url)
	if err != nil {
		return nil, err
	}

	return &ConcurrentHTTPObserver{
		observer: observer,
		wg:       &sync.WaitGroup{},
	}, nil
}

func (c *ConcurrentHTTPObserver) LogEvent(event Event) error {
	c.wg.Add(1)
	go func(e Event) {
		defer c.wg.Done()
		_ = c.observer.LogEvent(e)
	}(event)
	return nil
}

func (c *ConcurrentHTTPObserver) Wait() {
	c.wg.Wait()
}