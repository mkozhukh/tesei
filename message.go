package tesei

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Message is the fundamental unit of data that flows through the pipeline.
// It contains the data payload, a unique ID, metadata, and error state.
type Message[T any] struct {
	// ID is a unique identifier for the message.
	ID string
	// Data is the payload of the message.
	Data T
	// Metadata is a map for storing arbitrary key-value pairs.
	Metadata map[string]any

	// Error holds any error that occurred during processing of this message.
	Error error
	// ErrorStage indicates the stage where the error occurred.
	ErrorStage string
}

// NewMessage creates a new message with the given data and a generated ID.
func NewMessage[T any](data T) *Message[T] {
	return &Message[T]{
		ID:       generateID(),
		Data:     data,
		Metadata: make(map[string]any),
	}
}

// NewMessageWithID creates a new message with the given ID and data.
func NewMessageWithID[T any](id string, data *T) *Message[T] {
	return &Message[T]{
		ID:       id,
		Data:     *data,
		Metadata: make(map[string]any),
	}
}

// HasError returns true if the message contains an error.
func (m *Message[T]) HasError() bool {
	return m.Error != nil
}

// WithError sets the error and error stage on the message.
func (m *Message[T]) WithError(err error, stage string) *Message[T] {
	m.Error = err
	m.ErrorStage = stage
	return m
}

// Clone creates a shallow copy of the message.
// The Metadata map is copied, but the Data payload is shallow copied.
func (m *Message[T]) Clone() *Message[T] {
	n := Message[T]{
		ID:       m.ID,
		Data:     m.Data,
		Metadata: make(map[string]any),

		Error:      m.Error,
		ErrorStage: m.ErrorStage,
	}

	for k, v := range m.Metadata {
		n.Metadata[k] = v
	}

	return &n
}

func generateID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return time.Now().Format("20060102150405.000000")
	}
	return hex.EncodeToString(b)
}
