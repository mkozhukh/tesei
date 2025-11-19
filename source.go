package tesei

import (
	"fmt"
)

type End[T any] struct {
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

type Log[T any] struct {
	Message string
	Print   func(msg *Message[T], err error) string
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

type SetMetaData[T any] struct {
	Key     string
	Value   any
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
