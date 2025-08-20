package tesei

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Executor[T any] interface {
	Start(ctx context.Context) (time.Duration, error)
	Run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T])
	Input() chan<- *Message[T]
	Output() <-chan *Message[T]
}

type executor[T any] struct {
	stages     []stage[T]
	bufferSize int

	input  chan *Message[T]
	output chan *Message[T]
	cancel context.CancelFunc
}

func (e *executor[T]) Start(baseCtx context.Context) (time.Duration, error) {
	start := time.Now()
	base, cancel := context.WithCancel(baseCtx)
	ctx := NewThread(base, 1)
	e.cancel = cancel

	e.input = make(chan *Message[T], e.bufferSize)
	e.output = make(chan *Message[T], e.bufferSize)

	wg := sync.WaitGroup{}
	done := make(chan struct{})
	e.innerRun(ctx, &wg, done, e.input, e.output)

	select {
	case err := <-ctx.Error():
		e.cancel()
		return time.Since(start), fmt.Errorf("Executor error: %w", err)
	case <-ctx.Done():
		wg.Wait()
		return time.Since(start), ctx.Context.Err()
	case <-done:
		// All stages completed normally
		break
	}

	return time.Since(start), nil
}

func (e *executor[T]) Run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T]) {
	wg := sync.WaitGroup{}
	done := make(chan struct{})
	e.innerRun(ctx, &wg, done, in, out)

	select {
	case <-ctx.Done():
		wg.Wait()
		break
	case <-done:
		// All stages completed normally
		break
	}
}

func (e *executor[T]) innerRun(ctx *Thread, wg *sync.WaitGroup, done chan struct{}, globalIn <-chan *Message[T], globalOut chan<- *Message[T]) {
	if len(e.stages) == 0 {
		go func() {
			for range e.input {
			}
			close(e.output)
		}()
	}

	channels := e.wireChannels()

	for i, stg := range e.stages {
		wg.Add(1)
		var in <-chan *Message[T]
		if i == 0 {
			in = globalIn
		} else {
			in = channels[i]
		}

		var out chan<- *Message[T]
		if i == len(e.stages)-1 {
			out = globalOut
		} else {
			out = channels[i+1]
		}

		go func(s stage[T], input <-chan *Message[T], output chan<- *Message[T]) {
			s.run(ctx, input, output)
			wg.Done()
		}(stg, in, out)
	}

	go func() {
		wg.Wait()
		close(done)
	}()
}

func (e *executor[T]) Input() chan<- *Message[T] {
	return e.input
}

func (e *executor[T]) Output() <-chan *Message[T] {
	return e.output
}

func (e *executor[T]) wireChannels() []chan *Message[T] {
	channels := make([]chan *Message[T], len(e.stages)+1)

	for i := 1; i < len(channels)-1; i++ {
		channels[i] = make(chan *Message[T], e.bufferSize)
	}

	return channels
}
