package tesei

import (
	"context"
	"time"
)

func ExampleStringsSource() {
	p := NewPipeline[string]().
		Sequential(StringsSource{[]string{"hello", "world"}}).
		Sequential(End[string]{Log: true}).
		Build()

	ctx := context.Background()
	p.Start(ctx)

	// Output:
	// done: hello
	// done: world
}
func ExampleStringsSource_async() {
	p := NewPipeline[string]().
		Sequential(StringsSource{[]string{"hello", "world"}}).
		Sequential(End[string]{Log: true}).
		Build()

	ctx := context.Background()
	go p.Start(ctx)
	time.Sleep(10 * time.Millisecond)

	// Output:
	// done: hello
	// done: world
}
