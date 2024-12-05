package gchatmeow

import (
	"sync"
)

type Observer func(interface{})

// Event represents a callback function for channel events
type Event struct {
	observers []Observer
	mu        sync.RWMutex
}

// AddObserver adds an event observer
func (e *Event) AddObserver(observer Observer) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.observers = append(e.observers, observer)
}

// Fire triggers all observers for an event
func (e *Event) Fire(data interface{}) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, observer := range e.observers {
		go observer(data)
	}
}
