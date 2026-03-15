package events

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Event represents a system event that can be streamed to subscribers.
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	AppID     string                 `json:"app_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Filter specifies which events a subscriber wants to receive.
type Filter struct {
	EventType string // empty means all types
	AppID     string // empty means all apps
}

// subscriber holds a channel and its filter criteria.
type subscriber struct {
	id     string
	ch     chan Event
	filter Filter
}

// EventStream manages real-time event broadcasting with subscriber channels.
type EventStream struct {
	mu          sync.RWMutex
	subscribers map[string]*subscriber
	buffer      []Event
	bufferSize  int
}

// NewEventStream creates a new event streaming service.
func NewEventStream() *EventStream {
	return &EventStream{
		subscribers: make(map[string]*subscriber),
		buffer:      make([]Event, 0, 100),
		bufferSize:  100,
	}
}

// Subscribe creates a new subscription and returns the subscriber ID and event channel.
// The channel has a buffer of 64 events; slow consumers may miss events.
func (es *EventStream) Subscribe(filter Filter) (string, <-chan Event) {
	es.mu.Lock()
	defer es.mu.Unlock()

	id := uuid.NewString()
	ch := make(chan Event, 64)
	es.subscribers[id] = &subscriber{
		id:     id,
		ch:     ch,
		filter: filter,
	}
	return id, ch
}

// Unsubscribe removes a subscription and closes its channel.
func (es *EventStream) Unsubscribe(id string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	sub, ok := es.subscribers[id]
	if ok {
		close(sub.ch)
		delete(es.subscribers, id)
	}
}

// Publish broadcasts an event to all matching subscribers and stores it in the buffer.
func (es *EventStream) Publish(event Event) {
	if event.ID == "" {
		event.ID = uuid.NewString()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	es.mu.Lock()
	// Add to replay buffer.
	es.buffer = append(es.buffer, event)
	if len(es.buffer) > es.bufferSize {
		es.buffer = es.buffer[len(es.buffer)-es.bufferSize:]
	}

	// Collect matching subscribers while holding the lock.
	matching := make([]*subscriber, 0, len(es.subscribers))
	for _, sub := range es.subscribers {
		if matchesFilter(event, sub.filter) {
			matching = append(matching, sub)
		}
	}
	es.mu.Unlock()

	// Send to matching subscribers without holding the lock.
	for _, sub := range matching {
		select {
		case sub.ch <- event:
		default:
			// Subscriber channel full, skip to avoid blocking.
		}
	}
}

// RecentEvents returns the last N buffered events (up to bufferSize).
func (es *EventStream) RecentEvents() []Event {
	es.mu.RLock()
	defer es.mu.RUnlock()

	result := make([]Event, len(es.buffer))
	copy(result, es.buffer)
	return result
}

// RecentEventsFiltered returns recent events matching the given filter.
func (es *EventStream) RecentEventsFiltered(filter Filter) []Event {
	es.mu.RLock()
	defer es.mu.RUnlock()

	var result []Event
	for _, e := range es.buffer {
		if matchesFilter(e, filter) {
			result = append(result, e)
		}
	}
	if result == nil {
		result = []Event{}
	}
	return result
}

// matchesFilter checks if an event matches a subscriber's filter criteria.
func matchesFilter(event Event, filter Filter) bool {
	if filter.EventType != "" && event.Type != filter.EventType {
		return false
	}
	if filter.AppID != "" && event.AppID != filter.AppID {
		return false
	}
	return true
}
