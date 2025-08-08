package files

import (
	"context"
	"fmt"

	"github.com/mkozhukh/tesei"
)

func ExampleListDir() {
	err := tesei.NewPipeline[TextFile]().
		Sequential(ListDir{Path: "../testdata", Ext: ".txt"}).
		Sequential(WriteFile{DryRun: true}).
		Sequential(tesei.End[TextFile]{}).
		Build().
		Start(context.Background())

	if err != nil {
		fmt.Println("error:", err)
	}

	// Output:
	// write file: ../testdata/a.txt
	// write file: ../testdata/b.txt
}

func ExampleReadFile() {
	err := tesei.NewPipeline[TextFile]().
		Sequential(ListDir{Path: "../testdata", Ext: ".txt"}).
		Sequential(ReadFile{}).
		Sequential(tesei.TransformJob[TextFile]{
			Transform: func(msg *tesei.Message[TextFile]) (*tesei.Message[TextFile], error) {
				fmt.Println("file size:", msg.ID, len(msg.Data.Content))
				return msg, nil
			},
		}).
		Sequential(tesei.End[TextFile]{}).
		Build().
		Start(context.Background())

	if err != nil {
		fmt.Println("error:", err)
	}

	// Output:
	// file size: ../testdata/a.txt 5
	// file size: ../testdata/b.txt 5
}
