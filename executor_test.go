package tesei_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mkozhukh/tesei"
)

func TestExecutorRun(t *testing.T) {
	p := tesei.NewPipeline[string]().
		Sequential(&tesei.TransformJob[string]{
			Transform: func(msg *tesei.Message[string]) (*tesei.Message[string], error) {
				msg.Data = strings.ToUpper(msg.Data)
				return msg, nil
			},
		})

	exec := p.Build()

	ctx := context.Background()

	go exec.Start(ctx)
	time.Sleep(10 * time.Millisecond)

	exec.Input() <- tesei.NewMessage("hello")
	exec.Input() <- tesei.NewMessage("world")
	close(exec.Input())

	result1 := <-exec.Output()
	if result1.Data != "HELLO" {
		t.Errorf("Expected 'HELLO', got %v", result1.Data)
	}

	result2 := <-exec.Output()
	if result2.Data != "WORLD" {
		t.Errorf("Expected 'WORLD', got %v", result2.Data)
	}

	select {
	case _, ok := <-exec.Output():
		if ok {
			t.Error("Expected output to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected output to be closed")
	}
}

func TestExecutorEmptyPipeline(t *testing.T) {
	p := tesei.NewPipeline[int]()
	exec := p.Build()

	ctx := context.Background()

	_, err := exec.Start(ctx)
	if err != nil {
		t.Errorf("Expected no error for empty pipeline, got %v", err)
	}

	close(exec.Input())

	select {
	case _, ok := <-exec.Output():
		if ok {
			t.Error("Expected output to be closed for empty pipeline")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected output to be closed")
	}
}

func SkipTestExecutorContextCancellation(t *testing.T) {
	p := tesei.NewPipeline[int]().
		Sequential(&tesei.TransformJob[int]{
			Transform: func(msg *tesei.Message[int]) (*tesei.Message[int], error) {
				time.Sleep(100 * time.Millisecond)
				return msg, nil
			},
		})

	exec := p.Build()

	ctx, cancel := context.WithCancel(context.Background())
	go exec.Start(ctx)
	time.Sleep(10 * time.Millisecond)

	for i := 0; i < 10; i++ {
		exec.Input() <- tesei.NewMessage(i)
	}

	time.Sleep(50 * time.Millisecond)
	cancel()

	time.Sleep(200 * time.Millisecond)

	select {
	case exec.Input() <- tesei.NewMessage(999):
		t.Error("Expected input to be closed after cancellation")
	default:
	}
}

func TestExecutorChannelWiring(t *testing.T) {
	var mu sync.Mutex
	processOrder := []string{}

	job1 := &tesei.TransformJob[string]{
		Transform: func(msg *tesei.Message[string]) (*tesei.Message[string], error) {
			mu.Lock()
			processOrder = append(processOrder, "job1")
			mu.Unlock()
			msg.Data += "_1"
			return msg, nil
		},
	}

	job2 := &tesei.TransformJob[string]{
		Transform: func(msg *tesei.Message[string]) (*tesei.Message[string], error) {
			mu.Lock()
			processOrder = append(processOrder, "job2")
			mu.Unlock()
			msg.Data += "_2"
			return msg, nil
		},
	}

	job3 := &tesei.TransformJob[string]{
		Transform: func(msg *tesei.Message[string]) (*tesei.Message[string], error) {
			mu.Lock()
			processOrder = append(processOrder, "job3")
			mu.Unlock()
			msg.Data += "_3"
			return msg, nil
		},
	}

	p := tesei.NewPipeline[string]().
		Sequential(job1).
		Sequential(job2).
		Sequential(job3)

	exec := p.Build()

	ctx := context.Background()
	go exec.Start(ctx)
	time.Sleep(10 * time.Millisecond)

	exec.Input() <- tesei.NewMessage("test")
	close(exec.Input())

	result := <-exec.Output()
	if result.Data != "test_1_2_3" {
		t.Errorf("Expected 'test_1_2_3', got %v", result.Data)
	}

	mu.Lock()
	if len(processOrder) != 3 {
		t.Errorf("Expected 3 processing steps, got %d", len(processOrder))
	}
	mu.Unlock()
}

func TestExecutorErrorPropagation(t *testing.T) {
	failingJob := &tesei.TransformJob[string]{
		Transform: func(msg *tesei.Message[string]) (*tesei.Message[string], error) {
			return msg, errors.New("test error")
		},
	}

	p := tesei.NewPipeline[string]().Sequential(failingJob)
	exec := p.Build()

	ctx := context.Background()
	go exec.Start(ctx)
	time.Sleep(10 * time.Millisecond)

	exec.Input() <- tesei.NewMessage("test")
	close(exec.Input())

	result := <-exec.Output()
	if result.Error == nil {
		t.Error("Expected error to be propagated in message")
	}

	if result.Error.Error() != "test error" {
		t.Errorf("Expected 'test error', got %v", result.Error)
	}
}

func TestExecutorComplexPipeline(t *testing.T) {
	uppercase := &tesei.TransformJob[string]{
		Transform: func(msg *tesei.Message[string]) (*tesei.Message[string], error) {
			msg.Data = strings.ToUpper(msg.Data)
			return msg, nil
		},
	}

	addPrefix := &tesei.TransformJob[string]{
		Transform: func(msg *tesei.Message[string]) (*tesei.Message[string], error) {
			msg.Data = "PREFIX_" + msg.Data
			return msg, nil
		},
	}

	addSuffix := &tesei.TransformJob[string]{
		Transform: func(msg *tesei.Message[string]) (*tesei.Message[string], error) {
			msg.Data = msg.Data + "_SUFFIX"
			return msg, nil
		},
	}

	p := tesei.NewPipeline[string]().
		Sequential(uppercase).
		Parallel(addPrefix, addSuffix).
		WithBufferSize(10)

	exec := p.Build()

	ctx := context.Background()
	go exec.Start(ctx)
	time.Sleep(10 * time.Millisecond)

	exec.Input() <- tesei.NewMessage("hello")
	close(exec.Input())

	results := make(map[string]bool)
	for i := 0; i < 2; i++ {
		result := <-exec.Output()
		if result == nil {
			t.Fatal("Received nil message")
		}
		results[result.Data] = true
	}

	if !results["PREFIX_HELLO"] {
		t.Error("Expected 'PREFIX_HELLO' in results")
	}

	if !results["HELLO_SUFFIX"] {
		t.Error("Expected 'HELLO_SUFFIX' in results")
	}
}

func TestExecutorBufferSize(t *testing.T) {
	p := tesei.NewPipeline[int]().
		Sequential(&tesei.TransformJob[int]{
			Transform: func(msg *tesei.Message[int]) (*tesei.Message[int], error) {
				time.Sleep(10 * time.Millisecond)
				return msg, nil
			},
		}).
		WithBufferSize(5)

	exec := p.Build()

	// Cannot check private field bufferSize in black-box test
	// Instead verify behavior (non-blocking send)

	ctx := context.Background()
	go exec.Start(ctx)
	time.Sleep(10 * time.Millisecond)

	for i := 0; i < 5; i++ {
		select {
		case exec.Input() <- tesei.NewMessage(i):
		case <-time.After(10 * time.Millisecond):
			t.Errorf("Expected to be able to send %d messages without blocking", i+1)
		}
	}

	close(exec.Input())
}

func TestExecutorParralelPipelines(t *testing.T) {
	var count int32

	a := tesei.NewPipeline[int]().
		Sequential(tesei.CounterJob[int]{Count: &count}).Build()

	b := tesei.NewPipeline[int]().
		Sequential(tesei.CounterJob[int]{Count: &count}).Build()

	p := tesei.NewPipeline[int]().
		Sequential(tesei.Slice[int]{Items: []int{1, 2}}).
		Parallel(a, b).
		Build()

	p.Start(context.Background())

	if count != 4 {
		t.Errorf("Expected count to be 4, got %d", count)
	}
}

func TestExecutorSequentialPipelines(t *testing.T) {
	var count int32

	a := tesei.NewPipeline[int]().
		Sequential(tesei.CounterJob[int]{Count: &count}).Build()

	p := tesei.NewPipeline[int]().
		Sequential(tesei.Slice[int]{Items: []int{1, 2}}).
		Sequential(a).
		Sequential(tesei.End[int]{}).
		Build()

	p.Start(context.Background())

	if count != 2 {
		t.Errorf("Expected count to be 2, got %d", count)
	}
}
