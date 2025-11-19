package llm_test

import (
	"context"
	"fmt"

	echotemplates "github.com/mkozhukh/echo-templates"
	"github.com/mkozhukh/tesei"
	"github.com/mkozhukh/tesei/files"
	"github.com/mkozhukh/tesei/llm"
)

func ExampleCompleteContent() {

	llm.SetModel("mock/test")
	p := tesei.NewPipeline[files.TextFile]().
		Sequential(files.ListDir{Path: "../testdata", Ext: ".txt"}).
		Sequential(files.ReadFile{}).
		Sequential(llm.CompleteContent{}).
		Sequential(files.PrintContent{}).
		Sequential(tesei.End[files.TextFile]{}).
		Build()

	_, err := p.Start(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// ../testdata/a.txt
	// [user]: fileA
	// ../testdata/b.txt
	// [user]: fileB

}

func ExampleCompleteContent_withPrompt() {

	llm.SetModel("mock/test")
	p := tesei.NewPipeline[files.TextFile]().
		Sequential(files.ListDir{Path: "../testdata", Ext: ".txt"}).
		Sequential(files.ReadFile{}).
		Sequential(llm.CompleteContent{
			Prompt: "some",
		}).
		Sequential(files.PrintContent{}).
		Sequential(tesei.End[files.TextFile]{}).
		Build()

	_, err := p.Start(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// ../testdata/a.txt
	// [system]: some
	// [user]: fileA
	// ../testdata/b.txt
	// [system]: some
	// [user]: fileB

}

func ExampleCompleteTemplateString() {

	llm.SetModel("mock/test")
	p := tesei.NewPipeline[files.TextFile]().
		Sequential(files.ListDir{Path: "../testdata", Ext: ".txt"}).
		Sequential(files.ReadFile{}).
		Sequential(llm.CompleteTemplateString{
			Template: "@system: X\n@user: {{user_query}}",
		}).
		Sequential(files.PrintContent{}).
		Sequential(tesei.End[files.TextFile]{}).
		Build()

	_, err := p.Start(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// ../testdata/a.txt
	// [system]: X
	// [user]: fileA
	// ../testdata/b.txt
	// [system]: X
	// [user]: fileB

}

func ExampleCompleteTemplate() {

	source := echotemplates.NewMockSource(map[string]string{
		"do.md": "@system: X\n@user: {{user_query}} {{x|1}}",
	})

	llm.SetModel("mock/test")
	llm.SetTemplatesSource(source)

	p := tesei.NewPipeline[files.TextFile]().
		Sequential(files.ListDir{Path: "../testdata", Ext: ".txt"}).
		Sequential(files.ReadFile{}).
		Sequential(tesei.SetMetaData[files.TextFile]{
			Key:   "x",
			Value: 100,
		}).
		Sequential(llm.CompleteTemplate{
			Template: "do",
		}).
		Sequential(files.PrintContent{}).
		Sequential(tesei.End[files.TextFile]{}).
		Build()

	_, err := p.Start(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// ../testdata/a.txt
	// [system]: X
	// [user]: fileA 100
	// ../testdata/b.txt
	// [system]: X
	// [user]: fileB 100
}

func ExampleCompleteTemplate_withVars() {
	source := echotemplates.NewMockSource(map[string]string{
		"do.md": "@system: X\n@user: {{user_query}} {{x|1}} {{y|2}}",
	})

	llm.SetModel("mock/test")
	llm.SetTemplatesSource(source)

	p := tesei.NewPipeline[files.TextFile]().
		Sequential(files.ListDir{Path: "../testdata", Ext: ".txt"}).
		Sequential(files.ReadFile{}).
		Sequential(tesei.SetMetaData[files.TextFile]{
			Key:   "hash",
			Value: "123",
		}).
		Sequential(llm.CompleteTemplate{
			Template: "do",
			Vars: map[string]any{
				"x": 100,
				"y": "{{hash}}",
			},
		}).
		Sequential(files.PrintContent{}).
		Sequential(tesei.End[files.TextFile]{}).
		Build()

	_, err := p.Start(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// ../testdata/a.txt
	// [system]: X
	// [user]: fileA 100 123
	// ../testdata/b.txt
	// [system]: X
	// [user]: fileB 100 123
}
