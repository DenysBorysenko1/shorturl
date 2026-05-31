package audit

import "sync"

type Broadcaster struct {
	observers []Observer
	mu        sync.RWMutex
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		observers: make([]Observer, 0),
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
		_ = observer.LogEvent(event)
	}
}
