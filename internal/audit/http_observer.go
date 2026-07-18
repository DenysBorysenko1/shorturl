package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type RetryableHTTPClient struct {
	client     *http.Client
	maxRetries int
	backoff    time.Duration
}

func NewRetryableHTTPClient(maxRetries int, timeout time.Duration) *RetryableHTTPClient {
	return &RetryableHTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		maxRetries: maxRetries,
		backoff:    100 * time.Millisecond,
	}
}

func (r *RetryableHTTPClient) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		resp, err = r.client.Do(req)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		if resp != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}

		if attempt < r.maxRetries {
			time.Sleep(r.backoff * time.Duration(attempt+1))
		}

		if req.Body != nil {
			body, _ := json.Marshal(req.Context().Value("event"))
			req.Body = io.NopCloser(bytes.NewBuffer(body))
		}
	}

	if err != nil {
		return nil, fmt.Errorf("after %d retries: %w", r.maxRetries, err)
	}
	return resp, fmt.Errorf("after %d retries: unexpected status code: %d", r.maxRetries, resp.StatusCode)
}

type HTTPObserver struct {
	url    string
	client *RetryableHTTPClient
}

func NewHTTPObserver(url string) (*HTTPObserver, error) {
	if url == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	return &HTTPObserver{
		url:    url,
		client: NewRetryableHTTPClient(3, 5*time.Second),
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
	events   chan Event
	workers  int
	shutdown chan struct{}
}

func NewConcurrentHTTPObserver(url string) (*ConcurrentHTTPObserver, error) {
	observer, err := NewHTTPObserver(url)
	if err != nil {
		return nil, err
	}

	c := &ConcurrentHTTPObserver{
		observer: observer,
		wg:       &sync.WaitGroup{},
		events:   make(chan Event, 100),
		workers:  5,
		shutdown: make(chan struct{}),
	}

	c.startWorkers()
	return c, nil
}

func (c *ConcurrentHTTPObserver) startWorkers() {
	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			for {
				select {
				case event := <-c.events:
					_ = c.observer.LogEvent(event)
				case <-c.shutdown:
					return
				}
			}
		}()
	}
}

func (c *ConcurrentHTTPObserver) LogEvent(event Event) error {
	select {
	case c.events <- event:
		return nil
	default:
		return fmt.Errorf("event buffer is full, dropping event")
	}
}

func (c *ConcurrentHTTPObserver) Close() error {
	close(c.shutdown)
	c.wg.Wait()
	return nil
}
