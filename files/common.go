package files

import (
	"strings"

	"github.com/mkozhukh/tesei"
)

// Replace is a job that performs string replacements on the content of TextFile messages.
type Replace struct {
	// Matches is a map of strings to replace. Key is the target, Value is the replacement.
	// Value can contain template placeholders resolved against message metadata.
	Matches map[string]string
}

func (c Replace) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	tesei.Transform(ctx, in, out, func(msg *tesei.Message[TextFile]) (*tesei.Message[TextFile], error) {
		for k, v := range c.Matches {
			v := ResolveString(v, msg)
			msg.Data.Content = strings.ReplaceAll(msg.Data.Content, k, v)
		}
		return msg, nil
	})
}

// Filter is a job that filters TextFile messages based on a custom predicate.
type Filter struct {
	// Match is the predicate function. If it returns true, the message is passed through.
	Match func(msg *tesei.Message[TextFile]) bool
}

func (c Filter) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	tesei.Filter(ctx, in, out, c.Match)
}
