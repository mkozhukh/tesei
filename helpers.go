package tesei

import "sync/atomic"

// CounterJob is a job that counts the number of messages passing through it.
// It uses atomic operations to be safe for concurrent use.
type CounterJob[T any] struct {
	Count *int32
}

func (c CounterJob[T]) Run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T]) {
	Transform(ctx, in, out, func(msg *Message[T]) (*Message[T], error) {
		atomic.AddInt32(c.Count, 1)
		return msg, nil
	})
}
