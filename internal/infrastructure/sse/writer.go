package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Writer wraps an http.ResponseWriter for SSE streaming.
// Caller must set Content-Type: text/event-stream before first write.
type Writer struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// New returns a Writer if the ResponseWriter supports flushing.
// Returns an error if the underlying transport doesn't support SSE.
func New(w http.ResponseWriter) (*Writer, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("sse: ResponseWriter does not support flushing")
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering
	return &Writer{w: w, flusher: flusher}, nil
}

// WriteEvent sends an SSE event with a named event type and JSON data payload.
func (sw *Writer) WriteEvent(eventType string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("sse: marshal event data: %w", err)
	}
	_, err = fmt.Fprintf(sw.w, "event: %s\ndata: %s\n\n", eventType, string(b))
	if err != nil {
		return fmt.Errorf("sse: write event: %w", err)
	}
	sw.flusher.Flush()
	return nil
}

// WriteDelta sends a streaming text chunk.
func (sw *Writer) WriteDelta(delta string) error {
	return sw.WriteEvent("delta", map[string]string{"delta": delta})
}

// WriteDone sends the stream termination event.
func (sw *Writer) WriteDone(sessionID string, tokenCost int, latencyMs int64) error {
	return sw.WriteEvent("done", map[string]interface{}{
		"session_id": sessionID,
		"token_cost": tokenCost,
		"latency_ms": latencyMs,
	})
}

// WriteError sends an error event and signals stream termination.
// After calling WriteError, the connection should be closed.
func (sw *Writer) WriteError(code, message string) error {
	return sw.WriteEvent("error", map[string]string{
		"code":    code,
		"message": message,
	})
}
