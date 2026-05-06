package sse

import (
	"fmt"
	"net/http"
	"sync"
)

// Event, SSE üzerinden gönderilecek bir olayı temsil eder.
type Event struct {
	Type string
	Data string
}

// Broker, SSE client'larını yöneten yapıdır.
type Broker struct {
	mu      sync.RWMutex
	clients map[chan Event]struct{}
}

// NewBroker, yeni bir SSE broker oluşturur.
func NewBroker() *Broker {
	return &Broker{
		clients: make(map[chan Event]struct{}),
	}
}

// Subscribe, yeni bir client kaydeder ve event kanalı döner.
func (b *Broker) Subscribe() chan Event {
	ch := make(chan Event, 64)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

// Unsubscribe, bir client'ı kaydından siler.
func (b *Broker) Unsubscribe(ch chan Event) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
	close(ch)
}

// Publish, tüm bağlı client'lara event gönderir.
func (b *Broker) Publish(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- e:
		default:
			// Yavaş client — event atla
		}
	}
}

// ServeHTTP, SSE stream endpoint'i olarak çalışır.
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE desteklenmiyor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := b.Subscribe()
	defer b.Unsubscribe(ch)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			if evt.Type != "" {
				fmt.Fprintf(w, "event: %s\n", evt.Type)
			}
			fmt.Fprintf(w, "data: %s\n\n", evt.Data)
			flusher.Flush()
		}
	}
}
