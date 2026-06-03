package audit

import (
	"sync"

	"go.uber.org/zap"
)

type Broadcaster struct {
	observers []Observer
	logger    *zap.Logger
	mu        sync.RWMutex
}

func NewBroadcaster(logger *zap.Logger) *Broadcaster {
	return &Broadcaster{
		observers: make([]Observer, 0),
		logger:    logger,
	}
}

func (b *Broadcaster) Attach(observer Observer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.observers = append(b.observers, observer)
}

func (b *Broadcaster) Detach(observer Observer) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, obs := range b.observers {
		if obs == observer {
			b.observers = append(b.observers[:i], b.observers[i+1:]...)
			break
		}
	}
}

func (b *Broadcaster) Notify(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, observer := range b.observers {
		if err := observer.LogEvent(event); err != nil {
			b.logger.Error("observer failed to log event",
				zap.Error(err),
				zap.String("action", event.Action),
				zap.String("userID", event.UserID),
				zap.String("url", event.URL),
			)
		}
	}
}
