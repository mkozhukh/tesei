package tesei

import (
	"errors"
	"testing"
)

func TestNewMessage(t *testing.T) {
	data := "test data"
	msg := NewMessage(data)

	if msg.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if msg.Data != data {
		t.Errorf("Expected data to be %v, got %v", data, msg.Data)
	}

	if msg.Metadata == nil {
		t.Error("Expected metadata to be initialized")
	}

	if len(msg.Metadata) != 0 {
		t.Error("Expected metadata to be empty")
	}

	if msg.Error != nil {
		t.Error("Expected error to be nil")
	}

	if msg.ErrorStage != "" {
		t.Error("Expected error stage to be empty")
	}
}

func TestMessageIDUniqueness(t *testing.T) {
	msg1 := NewMessage("data1")
	msg2 := NewMessage("data2")

	if msg1.ID == msg2.ID {
		t.Error("Expected unique IDs for different messages")
	}
}

func TestMessageHasError(t *testing.T) {
	msg := NewMessage("test")

	if msg.HasError() {
		t.Error("Expected HasError to return false for message without error")
	}

	msg.Error = errors.New("test error")

	if !msg.HasError() {
		t.Error("Expected HasError to return true for message with error")
	}
}

func TestMessageWithError(t *testing.T) {
	msg := NewMessage("test")
	testErr := errors.New("test error")
	stage := "TestStage"

	result := msg.WithError(testErr, stage)

	if result != msg {
		t.Error("Expected WithError to return the same message pointer")
	}

	if msg.Error != testErr {
		t.Errorf("Expected error to be %v, got %v", testErr, msg.Error)
	}

	if msg.ErrorStage != stage {
		t.Errorf("Expected error stage to be %s, got %s", stage, msg.ErrorStage)
	}
}

func TestMessageMetadata(t *testing.T) {
	msg := NewMessage("test")

	msg.Metadata["key1"] = "value1"
	msg.Metadata["key2"] = "value2"

	if msg.Metadata["key1"] != "value1" {
		t.Errorf("Expected metadata key1 to be value1, got %s", msg.Metadata["key1"])
	}

	if msg.Metadata["key2"] != "value2" {
		t.Errorf("Expected metadata key2 to be value2, got %s", msg.Metadata["key2"])
	}

	if len(msg.Metadata) != 2 {
		t.Errorf("Expected metadata to have 2 entries, got %d", len(msg.Metadata))
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	if id1 == "" {
		t.Error("Expected generateID to return non-empty string")
	}

	if id1 == id2 {
		t.Error("Expected generateID to return unique IDs")
	}

	if len(id1) != 32 {
		t.Errorf("Expected ID to be 32 characters (hex encoding of 16 bytes), got %d", len(id1))
	}
}
