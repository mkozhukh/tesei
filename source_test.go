package tesei_test

import (
	"context"
	"fmt"
	"time"

	"github.com/mkozhukh/tesei"
)

func ExampleSlice_string() {
	p := tesei.NewPipeline[string]().
		Sequential(tesei.Slice[string]{Items: []string{"hello", "world"}}).
		Sequential(tesei.Log[string]{Print: func(msg *tesei.Message[string], err error) string {
			return "done: " + msg.Data
		}}).
		Sequential(tesei.End[string]{}).
		Build()

	ctx := context.Background()
	p.Start(ctx)
	fmt.Println("---")

	// Output:
	// done: hello
	// done: world
	// ---
}
func ExampleSlice_string_async() {
	p := tesei.NewPipeline[string]().
		Sequential(tesei.Slice[string]{Items: []string{"hello", "world"}}).
		Sequential(tesei.Log[string]{Print: func(msg *tesei.Message[string], err error) string {
			return "done: " + msg.Data
		}}).
		Build()

	ctx := context.Background()
	go p.Start(ctx)
	fmt.Println("---")
	time.Sleep(10 * time.Millisecond)

	// Output:
	// ---
	// done: hello
	// done: world
}
