package gchatmeow

import (
	"sync"
)

type Observer func(interface{})

type Event struct {
	observers []Observer
	mu        sync.RWMutex
}

func (e *Event) AddObserver(observer Observer) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.observers = append(e.observers, observer)
}

func (e *Event) Fire(data interface{}) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, observer := range e.observers {
		go observer(data)
	}
}
