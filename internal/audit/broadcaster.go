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

	go b.runObserverWorker(entry)
}

func (b *Broadcaster) runObserverWorker(entry observerEntry) {
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
