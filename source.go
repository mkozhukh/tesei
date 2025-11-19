package tesei

import (
	"fmt"
)

// End is a sink job that consumes all messages.
// It is required at the end of the pipeline to prevent blocking.
type End[T any] struct {
	// Log determines if the job should log the completion of each message.
	Log bool
}

func (e End[T]) Run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T]) {
	defer close(out)
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-in:
			if !ok {
				return
			}

			if e.Log {
				if msg.Error != nil {
					fmt.Println("error:", msg.ID, msg.Error)
				} else {
					fmt.Println("done:", msg.ID)
				}
			}
		}
	}
}

// Log is a job that logs messages as they pass through.
// It does not modify the messages.
type Log[T any] struct {
	// Message is a prefix for the log message.
	Message string
	// Print is a custom function to format the log message.
	Print func(msg *Message[T], err error) string
}

func (l Log[T]) Run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T]) {
	defer close(out)
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-in:
			if !ok {
				return
			}

			if l.Print != nil {
				fmt.Println(l.Print(msg, msg.Error))
			} else {
				if msg.Error != nil {
					errorStr := msg.Error.Error()
					if msg.ErrorStage != "" {
						errorStr = msg.ErrorStage + ": " + errorStr
					}
					fmt.Println("[error]", l.Message, msg.ID, errorStr)
				} else {
					fmt.Println("[ok]", l.Message, msg.ID)
				}
			}

			select {
			case out <- msg:
			case <-ctx.Done():
				return
			}
		}
	}
}

// SetMetaData is a job that sets a metadata key-value pair on passing messages.
type SetMetaData[T any] struct {
	// Key is the metadata key to set.
	Key string
	// Value is the value to set. Used if Handler is nil.
	Value any
	// Handler is a function to generate the value dynamically based on the message.
	Handler func(msg *Message[T]) any
}

func (s SetMetaData[T]) Run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T]) {
	Transform(ctx, in, out, func(msg *Message[T]) (*Message[T], error) {
		if s.Handler != nil {
			msg.Metadata[s.Key] = s.Handler(msg)
		} else {
			msg.Metadata[s.Key] = s.Value
		}
		return msg, nil
	})
}

// Slice is a source job that emits a slice of items as messages.
type Slice[T any] struct {
	Items []T
}

func (s Slice[T]) Run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T]) {
	defer close(out)
	for _, item := range s.Items {
		select {
		case out <- NewMessage(item):
		case <-ctx.Done():
			return
		}
	}
}
