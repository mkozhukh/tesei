package tesei

type Job[T any] interface {
	Run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T])
}

type JobFunc[T any] func(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T])

func (f JobFunc[T]) Run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T]) {
	f(ctx, in, out)
}

type TransformJob[T any] struct {
	ProcessError bool
	Transform    func(*Message[T]) (*Message[T], error)
}

func (t TransformJob[T]) Run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T]) {
	defer close(out)
	for {
		select {
		case msg, ok := <-in:
			if !ok {
				return
			}
			if msg.Error == nil || t.ProcessError {
				var err error
				msg, err = t.Transform(msg)
				if msg == nil {
					continue
				}
				if err != nil {
					msg.Error = err
				}
			}
			select {
			case out <- msg:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func Transform[T any](ctx *Thread, in <-chan *Message[T], out chan<- *Message[T], transform func(*Message[T]) (*Message[T], error)) {
	defer close(out)
	for {
		select {
		case msg, ok := <-in:
			if !ok {
				return
			}
			if msg.Error == nil {
				var err error
				msg, err = transform(msg)
				if msg == nil {
					continue
				}
				if err != nil {
					msg.Error = err
				}
			}
			select {
			case out <- msg:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func Filter[T any](ctx *Thread, in <-chan *Message[T], out chan<- *Message[T], filter func(*Message[T]) bool) {
	defer close(out)
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-in:
			if !ok {
				return
			}

			if filter(msg) {
				select {
				case out <- msg:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
