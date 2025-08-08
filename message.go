package tesei

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Message[T any] struct {
	ID       string
	Data     T
	Metadata map[string]any

	Error      error
	ErrorStage string
}

func NewMessage[T any](data T) *Message[T] {
	return &Message[T]{
		ID:       generateID(),
		Data:     data,
		Metadata: make(map[string]any),
	}
}

func NewMessageWithID[T any](id string, data *T) *Message[T] {
	return &Message[T]{
		ID:       id,
		Data:     *data,
		Metadata: make(map[string]any),
	}
}

func (m *Message[T]) HasError() bool {
	return m.Error != nil
}

func (m *Message[T]) WithError(err error, stage string) *Message[T] {
	m.Error = err
	m.ErrorStage = stage
	return m
}

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

// type TextFile struct {
// 	Name string
// 	Path string
// 	Content string
// }

// func NewTextFileMessage(path string) *TextFile {
// 	return &TextFile{
// 		Path: path,
// 	}
// }
