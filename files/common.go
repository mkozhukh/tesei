package files

import (
	"strings"

	"github.com/mkozhukh/tesei"
)

type Replace struct {
	Matches map[string]string
}

func (c Replace) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	tesei.Transform(ctx, in, out, func(msg *tesei.Message[TextFile]) (*tesei.Message[TextFile], error) {
		for k, v := range c.Matches {
			v := resolveString(v, msg)
			msg.Data.Content = strings.ReplaceAll(msg.Data.Content, k, v)
		}
		return msg, nil
	})
}

type Filter struct {
	Match func(msg *tesei.Message[TextFile]) bool
}

func (c Filter) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	tesei.Filter(ctx, in, out, c.Match)
}
