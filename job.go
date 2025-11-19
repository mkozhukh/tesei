package tesei

// Job is the interface for any processing unit in the pipeline.
// It reads messages from the input channel, processes them, and writes to the output channel.
type Job[T any] interface {
	Run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T])
}

// JobFunc is a function type that implements the Job interface.
type JobFunc[T any] func(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T])

func (f JobFunc[T]) Run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T]) {
	f(ctx, in, out)
}

// TransformJob is a helper struct for creating 1-to-1 transformation jobs.
// It handles the boilerplate of reading from input, checking for errors, and writing to output.
type TransformJob[T any] struct {
	// ProcessError determines if the job should process messages that already have an error.
	ProcessError bool
	// Transform is the function that processes the message.
	// If it returns nil, nil, the message is filtered out (consumed).
	Transform func(*Message[T]) (*Message[T], error)
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

// Transform is a helper function to create a transformation job from a function.
// It handles the boilerplate of reading from input, checking for errors, and writing to output.
// If the transform function returns nil, nil, the message is filtered out (consumed).
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

// Filter is a helper function to create a filtering job.
// It only passes messages for which the filter function returns true.
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
