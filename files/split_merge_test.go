package files

import (
	"context"
	"strings"
	"testing"

	"github.com/mkozhukh/tesei"
)

func TestSplitMerge(t *testing.T) {
	// Test data
	input := TextFile{
		Name:    "test.txt",
		Folder:  "/tmp",
		Content: "part1,part2,part3",
	}

	// Split logic
	splitter := Split{
		By: func(text string) []string {
			return strings.Split(text, ",")
		},
	}

	// Merge logic
	merger := Merge{
		Glue: "|",
	}

	// Capture output (we need a custom sink to verify, but for now we can rely on the fact that End consumes it)
	// To verify the output, let's use a custom job instead of End that stores the result
	var result *tesei.Message[TextFile]

	p := tesei.NewPipeline[TextFile]().
		Sequential(tesei.Slice[TextFile]{Items: []TextFile{input}}).
		Sequential(splitter).
		Sequential(merger).
		Sequential(tesei.TransformJob[TextFile]{
			Transform: func(msg *tesei.Message[TextFile]) (*tesei.Message[TextFile], error) {
				result = msg
				return msg, nil
			},
		}).
		Sequential(tesei.End[TextFile]{}).
		Build()

	_, err := p.Start(context.Background())
	if err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	expectedContent := "part1|part2|part3"
	if result.Data.Content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, result.Data.Content)
	}

	if result.ID == "" {
		t.Error("Expected non-empty ID")
	}
}

func TestSplitMetadata(t *testing.T) {
	input := TextFile{Content: "a,b"}
	splitter := Split{
		By: func(text string) []string { return strings.Split(text, ",") },
	}

	var chunks []*tesei.Message[TextFile]
	p := tesei.NewPipeline[TextFile]().
		Sequential(tesei.Slice[TextFile]{Items: []TextFile{input}}).
		Sequential(splitter).
		Sequential(tesei.TransformJob[TextFile]{
			Transform: func(msg *tesei.Message[TextFile]) (*tesei.Message[TextFile], error) {
				chunks = append(chunks, msg)
				return msg, nil
			},
		}).
		Sequential(tesei.End[TextFile]{}).
		Build()

	_, err := p.Start(context.Background())
	if err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	if len(chunks) != 2 {
		t.Fatalf("Expected 2 chunks, got %d", len(chunks))
	}

	if chunks[0].Metadata["split_index"] != 0 || chunks[0].Metadata["split_total"] != 2 {
		t.Errorf("Chunk 0 metadata incorrect: %v", chunks[0].Metadata)
	}
	if chunks[1].Metadata["split_index"] != 1 || chunks[1].Metadata["split_total"] != 2 {
		t.Errorf("Chunk 1 metadata incorrect: %v", chunks[1].Metadata)
	}
}

func TestClone(t *testing.T) {
	input := TextFile{
		Name:    "test.txt",
		Content: "original",
	}

	cloner := Clone{
		By: func(msg *tesei.Message[TextFile]) []*tesei.Message[TextFile] {
			m1 := msg.Clone()
			m1.Data.Name = "clone1.txt"
			m2 := msg.Clone()
			m2.Data.Name = "clone2.txt"
			return []*tesei.Message[TextFile]{m1, m2}
		},
	}

	var results []*tesei.Message[TextFile]

	p := tesei.NewPipeline[TextFile]().
		Sequential(tesei.Slice[TextFile]{Items: []TextFile{input}}).
		Sequential(cloner).
		Sequential(tesei.TransformJob[TextFile]{
			Transform: func(msg *tesei.Message[TextFile]) (*tesei.Message[TextFile], error) {
				results = append(results, msg)
				return msg, nil
			},
		}).
		Sequential(tesei.End[TextFile]{}).
		Build()

	_, err := p.Start(context.Background())
	if err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	if results[0].Data.Name != "clone1.txt" {
		t.Errorf("Expected clone1.txt, got %s", results[0].Data.Name)
	}
	if results[1].Data.Name != "clone2.txt" {
		t.Errorf("Expected clone2.txt, got %s", results[1].Data.Name)
	}
}
