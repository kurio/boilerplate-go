package ebus

import (
	"sync"

	boilerplater "github.com/kurio/boilerplate-go"
)

// Handler is the interface that wraps the basic Handle method.
//
// Handle will be invoked when a SystemEvent occurred.
type Handler interface {
	Handle(boilerplater.SystemEvent)
}

// HandlerFunc is the function adapter of Handler
type HandlerFunc func(boilerplater.SystemEvent)

// Handle handles the event. It invoke the f(e)
func (f HandlerFunc) Handle(e boilerplater.SystemEvent) {
	f(e)
}

// Bus is event bus implementation. It center of the event nerve system.
type Bus struct {
	mu       sync.RWMutex
	handlers []Handler
}

// Publish publishes a system event.
func (b *Bus) Publish(e boilerplater.SystemEvent) {
	b.mu.RLock()
	for _, h := range b.handlers {
		h.Handle(e)
	}
	b.mu.RUnlock()
}

// Subscribe register a handler to be called when a system event is being published.
func (b *Bus) Subscribe(h Handler) {
	b.mu.Lock()
	b.handlers = append(b.handlers, h)
	b.mu.Unlock()
}
