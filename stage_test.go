package tesei

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSequentialStage(t *testing.T) {
	job := JobFunc[string](func(ctx *Thread, in <-chan *Message[string], out chan<- *Message[string]) {
		defer close(out)
		for msg := range in {
			msg.Data = msg.Data + "_processed"
			out <- msg
		}
	})

	stage := &sequentialStage[string]{job: job}

	in := make(chan *Message[string], 1)
	out := make(chan *Message[string], 1)

	in <- NewMessage("test")
	close(in)

	ctx := NewThread(context.Background(), 1)
	stage.run(ctx, in, out)

	err := ctx.GetError()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	result := <-out
	if result.Data != "test_processed" {
		t.Errorf("Expected 'test_processed', got %v", result.Data)
	}
}

func TestParallelStage(t *testing.T) {
	var counter1, counter2 int32

	job1 := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
		for msg := range in {
			atomic.AddInt32(&counter1, 1)
			msg.Metadata["job"] = "job1"
			out <- msg
		}
		close(out)
	})

	job2 := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
		for msg := range in {
			atomic.AddInt32(&counter2, 1)
			msg.Metadata["job"] = "job2"
			out <- msg
		}
		close(out)
	})

	stage := &parallelStage[int]{jobs: []Job[int]{job1, job2}}

	in := make(chan *Message[int], 3)
	out := make(chan *Message[int], 6)

	for i := 0; i < 3; i++ {
		in <- NewMessage(i)
	}
	close(in)

	ctx := NewThread(context.Background(), 1)
	stage.run(ctx, in, out)

	err := ctx.GetError()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if atomic.LoadInt32(&counter1) != 3 {
		t.Errorf("Expected job1 to process 3 messages, got %d", counter1)
	}

	if atomic.LoadInt32(&counter2) != 3 {
		t.Errorf("Expected job2 to process 3 messages, got %d", counter2)
	}

	resultCount := 0
	job1Count := 0
	job2Count := 0

	for {
		select {
		case msg, ok := <-out:
			if !ok {
				goto done
			}
			resultCount++
			if msg.Metadata["job"] == "job1" {
				job1Count++
			} else if msg.Metadata["job"] == "job2" {
				job2Count++
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout waiting for results", resultCount)
			goto done
		}
	}

done:
	if resultCount != 6 {
		t.Errorf("Expected 6 results (3 messages * 2 jobs), got %d", resultCount)
	}

	if job1Count != 3 {
		t.Errorf("Expected 3 results from job1, got %d", job1Count)
	}

	if job2Count != 3 {
		t.Errorf("Expected 3 results from job2, got %d", job2Count)
	}
}

func TestFanOutStage(t *testing.T) {
	var counter int32
	var mu sync.Mutex
	processedBy := make(map[int]bool)

	job := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
		defer close(out)
		for msg := range in {
			atomic.AddInt32(&counter, 1)

			mu.Lock()
			processedBy[msg.Data] = true
			mu.Unlock()

			time.Sleep(10 * time.Millisecond)
			out <- msg
		}
	})

	stage := &fanOutStage[int]{job: job, count: 3}

	in := make(chan *Message[int], 10)
	out := make(chan *Message[int], 10)

	for i := 0; i < 10; i++ {
		in <- NewMessage(i)
	}
	close(in)

	ctx := NewThread(context.Background(), 1)

	start := time.Now()
	stage.run(ctx, in, out)
	duration := time.Since(start)

	err := ctx.GetError()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if atomic.LoadInt32(&counter) != 10 {
		t.Errorf("Expected 10 messages to be processed, got %d", counter)
	}

	mu.Lock()
	if len(processedBy) != 10 {
		t.Errorf("Expected all 10 messages to be processed, got %d unique messages", len(processedBy))
	}
	mu.Unlock()

	if duration > 150*time.Millisecond {
		t.Logf("Warning: Expected faster processing with 3 workers, took %v", duration)
	}

	resultCount := 0
	for {
		select {
		case _, ok := <-out:
			if !ok {
				goto done
			}
			resultCount++
		case <-time.After(100 * time.Millisecond):
			goto done
		}
	}

done:
	if resultCount != 10 {
		t.Errorf("Expected 10 results, got %d", resultCount)
	}
}

func TestParallelStageContextCancellation(t *testing.T) {
	var counter int32
	job1 := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
		defer close(out)
		for {
			select {
			case msg, ok := <-in:
				if !ok {
					return
				}
				atomic.AddInt32(&counter, 1)
				time.Sleep(100 * time.Millisecond)
				select {
				case out <- msg:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	})

	job2 := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
		defer close(out)
		for {
			select {
			case msg, ok := <-in:
				if !ok {
					return
				}
				atomic.AddInt32(&counter, 1)
				time.Sleep(100 * time.Millisecond)
				select {
				case out <- msg:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	})

	stage := &parallelStage[int]{jobs: []Job[int]{job1, job2}}

	in := make(chan *Message[int], 10)
	out := make(chan *Message[int], 20)

	for i := 0; i < 10; i++ {
		in <- NewMessage(i)
	}

	ctx, cancel := context.WithCancel(context.Background())
	thread := NewThread(ctx, 1)

	done := make(chan error)
	go func() {
		stage.run(thread, in, out)
		done <- nil
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()
	close(in)

	<-done // wait til stage is done
	err := thread.GetError()
	if err != nil && err != context.Canceled {
		t.Errorf("Expected nil or context.Canceled error, got %v", err)
	}

	if atomic.LoadInt32(&counter) >= 10 {
		t.Errorf("Expected less than 10 messages to be processed, got %d", counter)
	}
}

func TestFanOutStageClosesOutput(t *testing.T) {
	job := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
		for msg := range in {
			out <- msg
		}
		close(out)
	})

	stage := &fanOutStage[int]{job: job, count: 2}

	in := make(chan *Message[int])
	out := make(chan *Message[int], 1)

	ctx := NewThread(context.Background(), 1)

	done := make(chan bool)
	go func() {
		stage.run(ctx, in, out)
		done <- true
	}()

	close(in)

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected run to complete after input channel closed")
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
