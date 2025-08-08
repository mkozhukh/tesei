package tesei

import (
	"context"
	"sync"
)

type stage[T any] interface {
	run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T])
}

type sequentialStage[T any] struct {
	job Job[T]
}

func (s *sequentialStage[T]) run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T]) {
	s.job.Run(ctx, in, out)
}

type parallelStage[T any] struct {
	jobs []Job[T]
}

func (s *parallelStage[T]) run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T]) {
	inChannels := make([]chan *Message[T], len(s.jobs))
	outChannels := make([]chan *Message[T], len(s.jobs))
	for i := range inChannels {
		inChannels[i] = make(chan *Message[T], 1)
		outChannels[i] = make(chan *Message[T], 1)
	}

	go oneToMany(ctx, in, inChannels)
	go manyToOne(ctx, outChannels, out)

	var wg sync.WaitGroup

	for i, job := range s.jobs {
		wg.Add(1)
		go func(ind int, jb Job[T]) {
			defer wg.Done()
			jb.Run(ctx, inChannels[ind], outChannels[ind])
		}(i, job)
	}

	wg.Wait()
}

type fanOutStage[T any] struct {
	job   Job[T]
	count int
}

func (s *fanOutStage[T]) run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T]) {
	outChannels := make([]chan *Message[T], s.count)
	for i := range outChannels {
		outChannels[i] = make(chan *Message[T], 1)
	}

	go manyToOne(ctx, outChannels, out)
	var wg sync.WaitGroup

	for i := range s.count {
		wg.Add(1)
		go func(ind int, jb Job[T]) {
			defer wg.Done()
			jb.Run(ctx, in, outChannels[ind])
		}(i, s.job)
	}

	wg.Wait()
}

func oneToMany[T any](ctx context.Context, in <-chan *Message[T], out []chan *Message[T]) {
	defer func() {
		for _, ch := range out {
			if ch != nil {
				close(ch)
			}
		}
	}()

	// Read from input and forward a cloned message to each output channel
	for {
		select {
		case <-ctx.Done():
			// Context canceled; stop routing and close all outputs
			return
		case msg, ok := <-in:
			if !ok {
				// Input closed; close all outputs and exit
				return
			}
			// Fan out cloned message to all outputs with ctx awareness
			for _, ch := range out {
				if ch == nil {
					continue
				}
				cloned := msg.Clone()
				select {
				case <-ctx.Done():
					return
				case ch <- cloned:
				}
			}
		}
	}
}

func manyToOne[T any](ctx context.Context, ins []chan *Message[T], out chan<- *Message[T]) {
	var wg sync.WaitGroup
	for _, ch := range ins {
		if ch == nil {
			continue
		}
		wg.Add(1)
		go func(c <-chan *Message[T]) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					// Context canceled; stop routing and close all outputs
					return
				case msg, ok := <-c:
					if !ok {
						// Input closed; close all outputs and exit
						return
					}
					select {
					case <-ctx.Done():
						// Context canceled; stop routing and close all outputs
						return
					case out <- msg:
					}
				}
			}
		}(ch)
	}
	wg.Wait()
	close(out)
}
