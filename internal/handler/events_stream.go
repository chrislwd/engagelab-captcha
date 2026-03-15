package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/engagelab/captcha/internal/service/events"
)

// EventStreamHandler handles SSE streaming and recent events endpoints.
type EventStreamHandler struct {
	stream *events.EventStream
}

// NewEventStreamHandler creates a new EventStreamHandler.
func NewEventStreamHandler(stream *events.EventStream) *EventStreamHandler {
	return &EventStreamHandler{stream: stream}
}

// Stream handles GET /v1/events/stream (Server-Sent Events).
func (h *EventStreamHandler) Stream(c *gin.Context) {
	eventType := c.Query("event_type")
	appID := c.Query("app_id")

	filter := events.Filter{
		EventType: eventType,
		AppID:     appID,
	}

	subID, ch := h.stream.Subscribe(filter)
	defer h.stream.Unsubscribe(subID)

	// Set SSE headers.
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	// If requested, replay recent events first.
	if c.Query("replay") == "true" {
		recent := h.stream.RecentEventsFiltered(filter)
		for _, event := range recent {
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(c.Writer, "id: %s\nevent: %s\ndata: %s\n\n", event.ID, event.Type, string(data))
		}
		c.Writer.(http.Flusher).Flush()
	}

	// Stream events until the client disconnects.
	clientGone := c.Request.Context().Done()
	for {
		select {
		case <-clientGone:
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			_, writeErr := fmt.Fprintf(c.Writer, "id: %s\nevent: %s\ndata: %s\n\n", event.ID, event.Type, string(data))
			if writeErr != nil {
				return
			}
			if f, ok := c.Writer.(io.Closer); ok {
				_ = f
			}
			c.Writer.(http.Flusher).Flush()
		}
	}
}

// Recent handles GET /v1/events/recent.
func (h *EventStreamHandler) Recent(c *gin.Context) {
	eventType := c.Query("event_type")
	appID := c.Query("app_id")

	filter := events.Filter{
		EventType: eventType,
		AppID:     appID,
	}

	var recentEvents []events.Event
	if eventType != "" || appID != "" {
		recentEvents = h.stream.RecentEventsFiltered(filter)
	} else {
		recentEvents = h.stream.RecentEvents()
	}

	c.JSON(http.StatusOK, gin.H{
		"events": recentEvents,
		"count":  len(recentEvents),
	})
}
