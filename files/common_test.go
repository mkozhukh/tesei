package files

import (
	"context"
	"fmt"

	"github.com/mkozhukh/tesei"
)

func ExampleReplace_withVars() {
	p := tesei.NewPipeline[TextFile]().
		Sequential(Source{
			Files: []TextFile{
				{Name: "fileA", Content: "fileA"},
				{Name: "fileB", Content: "fileB"},
			},
		}).
		Sequential(Replace{
			Matches: map[string]string{
				"file": "none",
			},
		}).
		Sequential(PrintContent{}).
		Sequential(tesei.End[TextFile]{}).
		Build()

	_, err := p.Start(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// fileA
	// noneA
	// fileB
	// noneB
}
