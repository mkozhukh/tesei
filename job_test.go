package tesei

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestJobFuncAdapter(t *testing.T) {

	called := false
	jobFunc := JobFunc[string](func(ctx *Thread, in <-chan *Message[string], out chan<- *Message[string]) {
		called = true
		defer close(out)
		for msg := range in {
			out <- msg
		}
	})

	in := make(chan *Message[string], 1)
	out := make(chan *Message[string], 1)

	msg := NewMessage("test")
	in <- msg
	close(in)

	ctx := NewThread(context.Background(), 10)
	jobFunc.Run(ctx, in, out)

	err := ctx.GetError()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !called {
		t.Error("Expected JobFunc to be called")
	}

	result := <-out
	if result.Data != "test" {
		t.Errorf("Expected data to be 'test', got %v", result.Data)
	}
}

func TestTransformJob(t *testing.T) {
	transform := &TransformJob[string]{
		Transform: func(msg *Message[string]) (*Message[string], error) {
			str := msg.Data
			msg.Data = strings.ToUpper(str)
			return msg, nil
		},
	}

	in := make(chan *Message[string], 2)
	out := make(chan *Message[string], 2)

	in <- NewMessage("hello")
	in <- NewMessage("world")
	close(in)

	ctx := NewThread(context.Background(), 10)
	transform.Run(ctx, in, out)

	err := ctx.GetError()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	result1 := <-out
	if result1.Data != "HELLO" {
		t.Errorf("Expected 'HELLO', got %v", result1.Data)
	}

	result2 := <-out
	if result2.Data != "WORLD" {
		t.Errorf("Expected 'WORLD', got %v", result2.Data)
	}

	select {
	case _, ok := <-out:
		if ok {
			t.Error("Expected output channel to be closed")
		}
	default:
		t.Error("Expected output channel to be closed")
	}
}

func TestTransformJobWithError(t *testing.T) {
	transform := &TransformJob[string]{
		Transform: func(msg *Message[string]) (*Message[string], error) {
			return msg, errors.New("transform error")
		},
	}

	in := make(chan *Message[string], 1)
	out := make(chan *Message[string], 1)

	in <- NewMessage("test")
	close(in)

	ctx := NewThread(context.Background(), 10)
	transform.Run(ctx, in, out)

	if ctx.GetError() != nil {
		t.Errorf("Expected no error from Run, got %v", ctx.GetError())
	}

	result := <-out
	if result.Error == nil {
		t.Error("Expected error in message")
	}

	if result.Error.Error() != "transform error" {
		t.Errorf("Expected 'transform error', got %v", result.Error)
	}
}

func TestTransformJobContextCancellation(t *testing.T) {
	counter := 0
	transform := &TransformJob[int]{
		Transform: func(msg *Message[int]) (*Message[int], error) {
			counter++
			time.Sleep(100 * time.Millisecond)
			return msg, nil
		},
	}

	in := make(chan *Message[int], 10)
	out := make(chan *Message[int], 10)

	for i := 0; i < 10; i++ {
		in <- NewMessage(i)
	}
	close(in)

	base, cancel := context.WithCancel(context.Background())
	ctx := NewThread(base, 10)

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	transform.Run(ctx, in, out)

	if counter != 1 {
		t.Errorf("Expected 1 message to be processed, got %d", counter)
	}
}

func TestTransformJobClosesOutput(t *testing.T) {
	transform := &TransformJob[int]{
		Transform: func(msg *Message[int]) (*Message[int], error) {
			return msg, nil
		},
	}

	in := make(chan *Message[int])
	out := make(chan *Message[int], 1)

	ctx := NewThread(context.Background(), 10)

	done := make(chan bool)
	go func() {
		transform.Run(ctx, in, out)
		done <- true
	}()

	close(in)

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected Run to complete after input channel closed")
	}

	select {
	case _, ok := <-out:
		if ok {
			t.Error("Expected output channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected output channel to be closed")
	}
}
