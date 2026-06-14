// Package audit provides event broadcasting and observation capabilities
// for tracking and logging application events.
// It supports multiple observers (file, HTTP, etc.) that receive
// events asynchronously through a broadcaster pattern.
package audit

import (
	"sync"

	"go.uber.org/zap"
)

const observerEventBuffer = 100

type observerEntry struct {
	observer Observer
	events   chan Event
}

type Broadcaster struct {
	entries []observerEntry
	logger  *zap.Logger
	mu      sync.RWMutex
	closed  bool
	wg      sync.WaitGroup
}

func NewBroadcaster(logger *zap.Logger) *Broadcaster {
	return &Broadcaster{
		entries: make([]observerEntry, 0),
		logger:  logger,
	}
}

func (b *Broadcaster) Attach(observer Observer) {
	b.mu.Lock()
	defer b.mu.Unlock()

	events := make(chan Event, observerEventBuffer)
	entry := observerEntry{observer: observer, events: events}
	b.entries = append(b.entries, entry)

	b.wg.Add(1)
	go b.runObserverWorker(entry)
}

func (b *Broadcaster) runObserverWorker(entry observerEntry) {
	defer b.wg.Done()
	for event := range entry.events {
		if err := entry.observer.LogEvent(event); err != nil {
			b.logger.Error("observer failed to log event",
				zap.Error(err),
				zap.String("action", event.Action),
				zap.String("userID", event.UserID),
				zap.String("url", event.URL),
			)
		}
	}
}

func (b *Broadcaster) Detach(observer Observer) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, entry := range b.entries {
		if entry.observer == observer {
			close(entry.events)
			b.entries = append(b.entries[:i], b.entries[i+1:]...)
			break
		}
	}
}

func (b *Broadcaster) Notify(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	for _, entry := range b.entries {
		go func(e observerEntry) {
			select {
			case e.events <- event:
			default:
				b.logger.Warn("dropping audit event, observer buffer is full",
					zap.String("action", event.Action),
					zap.String("userID", event.UserID),
					zap.String("url", event.URL),
				)
			}
		}(entry)
	}
}

func (b *Broadcaster) Close() {
	b.mu.Lock()
	b.closed = true
	entries := make([]observerEntry, len(b.entries))
	copy(entries, b.entries)
	b.mu.Unlock()

	for _, entry := range entries {
		close(entry.events)
	}

	b.wg.Wait()

	b.mu.Lock()
	b.entries = nil
	b.mu.Unlock()
}
