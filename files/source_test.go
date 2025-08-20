package files

import (
	"context"
	"fmt"

	"github.com/mkozhukh/tesei"
)

func ExampleListDir() {
	_, err := tesei.NewPipeline[TextFile]().
		Sequential(ListDir{Path: "../testdata", Ext: ".txt"}).
		Sequential(WriteFile{DryRun: true, Log: true}).
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
	_, err := tesei.NewPipeline[TextFile]().
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

func ExampleRenameFile() {
	_, err := tesei.NewPipeline[TextFile]().
		Sequential(ListDir{Path: "../testdata", Ext: ".txt"}).
		Sequential(RenameFile{Suffix: "_test"}).
		Sequential(WriteFile{DryRun: true, Log: true}).
		Sequential(tesei.End[TextFile]{}).
		Build().
		Start(context.Background())

	if err != nil {
		fmt.Println("error:", err)
	}

	// Output:
	// write file: ../testdata/a_test.txt
	// write file: ../testdata/b_test.txt
}

func ExampleRenameFile_withHash() {
	_, err := tesei.NewPipeline[TextFile]().
		Sequential(ListDir{Path: "../testdata", Ext: ".txt"}).
		Sequential(HashContent{Key: "hash", Size: 8}).
		Sequential(RenameFile{Suffix: "_{{hash}}"}).
		Sequential(WriteFile{DryRun: true, Log: true}).
		Sequential(tesei.End[TextFile]{}).
		Build().
		Start(context.Background())

	if err != nil {
		fmt.Println("error:", err)
	}

	// Output:
	// write file: ../testdata/a_ivgFrYaM.txt
	// write file: ../testdata/b_ivgFrYaM.txt
}

func ExampleRenameFile_withHashParralel() {
	_, err := tesei.NewPipeline[TextFile]().
		Sequential(ListDir{Path: "../testdata", Ext: ".txt", Limit: 1}).
		Sequential(HashContent{Key: "hash", Size: 8}).
		Parallel(
			RenameFile{Suffix: "_{{hash}}", Ext: ".js"},
			RenameFile{Suffix: "_{{hash}}", Ext: ".css"},
		).
		Sequential(WriteFile{DryRun: true, Log: true}).
		Sequential(tesei.End[TextFile]{}).
		Build().
		Start(context.Background())

	if err != nil {
		fmt.Println("error:", err)
	}

	// Output:
	// write file: ../testdata/a_ivgFrYaM.css
	// write file: ../testdata/a_ivgFrYaM.js
}

func ExampleRenameFile_withHashParralelPipelines() {
	js := tesei.NewPipeline[TextFile]().
		Sequential(RenameFile{Suffix: "_{{hash}}", Ext: ".js"}).
		Build()

	css := tesei.NewPipeline[TextFile]().
		Sequential(RenameFile{Suffix: "_{{hash}}", Ext: ".css"}).
		Build()

	_, err := tesei.NewPipeline[TextFile]().
		Sequential(ListDir{Path: "../testdata", Ext: ".txt"}).
		Sequential(HashContent{Key: "hash", Size: 8}).
		Parallel(js, css).
		Sequential(WriteFile{DryRun: true, Log: true}).
		Sequential(tesei.End[TextFile]{}).
		Build().
		Start(context.Background())

	if err != nil {
		fmt.Println("error:", err)
	}

	// Output:
	// write file: ../testdata/a_ivgFrYaM.css
	// write file: ../testdata/b_ivgFrYaM.css
	// write file: ../testdata/a_ivgFrYaM.js
	// write file: ../testdata/b_ivgFrYaM.js
}
