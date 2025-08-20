package files

import (
	"context"
	"fmt"
	"testing"

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

func TestCleanAfterLLM_cleanText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Replace right arrow with ->",
			input:    "Step 1 → Step 2 → Step 3",
			expected: "Step 1 -> Step 2 -> Step 3",
		},
		{
			name:     "Replace various arrow types",
			input:    "A → B ⟶ C ⇒ D ➔ E ➜ F ➡ G ⇨ H ⟹ I",
			expected: "A -> B -> C -> D -> E -> F -> G -> H -> I",
		},
		{
			name:     "Replace em dash with hyphen",
			input:    "This is a test — with em dash",
			expected: "This is a test - with em dash",
		},
		{
			name:     "Replace en dash with hyphen",
			input:    "Pages 10–20 are important",
			expected: "Pages 10-20 are important",
		},
		{
			name:     "Replace various dash types",
			input:    "Em—dash, en–dash, horizontal―bar, figure‒dash",
			expected: "Em-dash, en-dash, horizontal-bar, figure-dash",
		},
		{
			name:     "Replace non-breaking space",
			input:    "Hello\u00A0world with non-breaking space",
			expected: "Hello world with non-breaking space",
		},
		{
			name:     "Replace various Unicode spaces",
			input:    "En\u2002space, em\u2003space, thin\u2009space",
			expected: "En space, em space, thin space",
		},
		{
			name:     "Remove zero-width characters",
			input:    "Text\u200Bwith\u200Czero\u200Dwidth\uFEFFcharacters",
			expected: "Textwithzerowidthcharacters",
		},
		{
			name:     "Complex text with multiple replacements",
			input:    "Step 1 → Process data—clean\u00A0it\u2003and then ⇒ output",
			expected: "Step 1 -> Process data-clean it and then -> output",
		},
		{
			name:     "Text with ideographic space",
			input:    "Japanese　text　with　ideographic　spaces",
			expected: "Japanese text with ideographic spaces",
		},
		{
			name:     "Normal text unchanged",
			input:    "This is normal text with regular spaces and hyphens - no changes",
			expected: "This is normal text with regular spaces and hyphens - no changes",
		},
		{
			name:     "Mixed arrows and dashes",
			input:    "Input → Process — Output ⟶ Result",
			expected: "Input -> Process - Output -> Result",
		},
		{
			name:     "Text with figure space and punctuation space",
			input:    "Number\u2007123 and punct\u2008space",
			expected: "Number 123 and punct space",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only special characters",
			input:    "→—\u00A0\u200B",
			expected: "->- ",
		},
		{
			name:     "Multiple consecutive special spaces",
			input:    "Word\u00A0\u2002\u2003\u2009word",
			expected: "Word    word",
		},
		{
			name:     "Two-em and three-em dashes",
			input:    "Two⸺em and three⸻em dashes",
			expected: "Two-em and three-em dashes",
		},
		{
			name:     "Narrow no-break space and medium mathematical space",
			input:    "Math\u202Fformula\u205Fhere",
			expected: "Math formula here",
		},
		{
			name:     "Ogham space mark",
			input:    "Old\u1680Irish\u1680text",
			expected: "Old Irish text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := CleanAfterLLM{}
			result := cleaner.cleanText(tt.input)
			if result != tt.expected {
				t.Errorf("cleanText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCleanAfterLLM_Run(t *testing.T) {
	// Create a test message
	in := make(chan *tesei.Message[TextFile], 1)
	out := make(chan *tesei.Message[TextFile], 1)

	testContent := "Step 1 → Process—clean\u00A0data\u200Band → output"
	expectedContent := "Step 1 -> Process-clean dataand -> output"

	msg := &tesei.Message[TextFile]{
		Data: TextFile{
			Name:    "test.txt",
			Folder:  "/test",
			Content: testContent,
		},
	}

	in <- msg
	close(in)

	cleaner := CleanAfterLLM{}
	ctx := tesei.NewThread(context.Background(), 10)

	// Run in a goroutine since it processes channels
	go cleaner.Run(ctx, in, out)

	// Read the result
	result := <-out

	if result.Data.Content != expectedContent {
		t.Errorf("Run() transformed content = %q, want %q", result.Data.Content, expectedContent)
	}
}
