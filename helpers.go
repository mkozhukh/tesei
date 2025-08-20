package tesei

import "sync/atomic"

type counterJob[T any] struct {
	Count *int32
}

func (c counterJob[T]) Run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T]) {
	Transform(ctx, in, out, func(msg *Message[T]) (*Message[T], error) {
		atomic.AddInt32(c.Count, 1)
		return msg, nil
	})
}
