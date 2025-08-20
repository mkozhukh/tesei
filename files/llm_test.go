package files

import (
	"context"
	"fmt"

	echotemplates "github.com/mkozhukh/echo-templates"
	"github.com/mkozhukh/tesei"
)

func ExampleCompleteContent() {

	SetModel("mock/test")
	p := tesei.NewPipeline[TextFile]().
		Sequential(ListDir{Path: "../testdata", Ext: ".txt"}).
		Sequential(ReadFile{}).
		Sequential(CompleteContent{}).
		Sequential(PrintContent{}).
		Sequential(tesei.End[TextFile]{}).
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

	SetModel("mock/test")
	p := tesei.NewPipeline[TextFile]().
		Sequential(ListDir{Path: "../testdata", Ext: ".txt"}).
		Sequential(ReadFile{}).
		Sequential(CompleteContent{
			Prompt: "some",
		}).
		Sequential(PrintContent{}).
		Sequential(tesei.End[TextFile]{}).
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

	SetModel("mock/test")
	p := tesei.NewPipeline[TextFile]().
		Sequential(ListDir{Path: "../testdata", Ext: ".txt"}).
		Sequential(ReadFile{}).
		Sequential(CompleteTemplateString{
			Template: "@system: X\n@user: {{user_query}}",
		}).
		Sequential(PrintContent{}).
		Sequential(tesei.End[TextFile]{}).
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

	SetModel("mock/test")
	SetTemplatesSource(source)

	p := tesei.NewPipeline[TextFile]().
		Sequential(ListDir{Path: "../testdata", Ext: ".txt"}).
		Sequential(ReadFile{}).
		Sequential(tesei.SetMetaData[TextFile]{
			Key:   "x",
			Value: 100,
		}).
		Sequential(CompleteTemplate{
			Template: "do",
		}).
		Sequential(PrintContent{}).
		Sequential(tesei.End[TextFile]{}).
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

	SetModel("mock/test")
	SetTemplatesSource(source)

	p := tesei.NewPipeline[TextFile]().
		Sequential(ListDir{Path: "../testdata", Ext: ".txt"}).
		Sequential(ReadFile{}).
		Sequential(tesei.SetMetaData[TextFile]{
			Key:   "hash",
			Value: "123",
		}).
		Sequential(CompleteTemplate{
			Template: "do",
			Vars: map[string]any{
				"x": 100,
				"y": "{{hash}}",
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
	// ../testdata/a.txt
	// [system]: X
	// [user]: fileA 100 123
	// ../testdata/b.txt
	// [system]: X
	// [user]: fileB 100 123
}
