# Tesei

A Go library for building robust data processing pipelines.

## Why Tesei?

Tesei was born from the need to automate mass file processing, specifically for LLM-based workflows. It provides a type-safe, generic way to build pipelines that can handle sequential, parallel, and fan-out execution patterns. Whether you're processing thousands of text files, transforming data streams, or orchestrating complex LLM interactions, Tesei offers a clean and maintainable approach.

Key features:
- **Type-Safe**: Built with Go generics for compile-time safety.
- **Flexible**: Support for Sequential, Parallel, and Fan-Out execution models.
- **Simple**: Minimal API surface, easy to understand and extend.

## Installation

```bash
go get github.com/mkozhukh/tesei
```

## Standard Libraries

Tesei comes with a set of standard libraries for common tasks:

- **[files](files/README.md)**: File system operations (Read, Write, List) and text file processing.
- **[llm](llm/README.md)**: Integration with Large Language Models (OpenAI, Anthropic, etc.).
- **[text](text/README.md)**: Text processing and cleaning utilities (Markdown, LLM cleanup).

## Usage

### Basic Pipeline

```go
package main

import (
	"context"
	"fmt"

	"github.com/mkozhukh/tesei"
)

func main() {
	p := tesei.NewPipeline[string]().
		Sequential(tesei.Slice[string]{Items: []string{"a", "b", "c"}}).
		Sequential(tesei.TransformJob[string]{
			Transform: func(msg *tesei.Message[string]) (*tesei.Message[string], error) {
				msg.Data = msg.Data + "_processed"
				return msg, nil
			},
		}).
		Sequential(tesei.End[string]{}).
		Build()

	_, err := p.Start(context.Background())
	if err != nil {
		panic(err)
	}
}
```

### File Processing with LLM

```go
package main

import (
	"context"
	"os"

	"github.com/mkozhukh/tesei"
	"github.com/mkozhukh/tesei/files"
	"github.com/mkozhukh/tesei/llm"
)

func main() {
	llm.SetModel("openai/gpt-4o")

	p := tesei.NewPipeline[files.TextFile]().
		Sequential(files.ListDir{Path: "./docs", Ext: ".md"}).
		Sequential(files.ReadFile{}).
		Sequential(llm.CompleteContent{
			Prompt: "Summarize this document in one sentence.",
		}).
		Sequential(files.WriteFile{Folder: "./summaries"}).
		Sequential(tesei.End[files.TextFile]{}).
		Build()

	_, err := p.Start(context.Background())
	if err != nil {
		panic(err)
	}
}
```

## API Reference

For full details, please refer to the [GoDoc documentation](https://pkg.go.dev/github.com/mkozhukh/tesei).

### Pipeline Builder
- `NewPipeline[T]()`: Creates a new pipeline builder for type `T`.
- `Sequential(jobs ...Job[T])`: Adds one or more jobs to be executed sequentially.
- `Parallel(jobs ...Job[T])`: Adds a stage where input messages are broadcast to multiple jobs running in parallel.
- `FanOut(job Job[T], count int)`: Adds a stage where a single job is run by multiple workers (competing consumers).
- `WithBufferSize(size int)`: Sets the buffer size for channels between stages.
- `Build()`: Compiles the pipeline and returns an `Executor`.

### Core Interfaces
- `Job[T]`: The interface for any processing unit.
  ```go
  type Job[T any] interface {
      Run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T])
  }
  ```
- `Message[T]`: The data unit flowing through the pipeline. Contains `Data`, `ID`, `Metadata`, and `Error`.
- `Executor[T]`: The runtime engine created by `Build()`. Use `Start(ctx)` to run it.
  - **Note**: `Executor[T]` also implements `Job[T]`, so you can use a built pipeline as a job within another pipeline.

> [!IMPORTANT]
> **Mandatory End Job**: Top-level pipelines MUST end with a consumer job like `tesei.End[T]`. This job ensures all messages are pulled through the pipeline. Without it, the pipeline will block indefinitely once internal buffers are full.

### Helpers
- `TransformJob[T]`: A struct-based helper for simple 1-to-1 transformations.
- `Transform[T]`: A function helper to implement custom jobs without writing the loop/select boilerplate.

### Common jobs
- `Slice[T]`: A function helper to create a job that emits a slice of data.
- `Filter[T]`: A function helper to filter messages based on a predicate.
- `Log[T]`: A function helper to log messages.
- `End[T]`: A function helper to end the pipeline.

## Common Scenarios

### 1. Mass File Processing (Sequential)
**Scenario**: You have a directory of text files and want to read them, perform some operation, and write them back.

```go
tesei.NewPipeline[files.TextFile]().
    Sequential(files.ListDir{Path: "./input", Ext: ".txt"}). // 1. List files
    Sequential(files.ReadFile{}).                            // 2. Read content
    Sequential(myCustomProcessingJob).                       // 3. Modify content
    Sequential(files.WriteFile{DryRun: false}).              // 4. Save changes
    Build()
```

### 2. Parallel Analysis (Broadcasting)
**Scenario**: You want to perform multiple *different* analyses on the same file simultaneously (e.g., generate a summary AND extract keywords).

```go
// Define specific jobs
summarizeJob := files.CompleteContent{Prompt: "Summarize this text"}
keywordsJob := files.CompleteContent{Prompt: "Extract keywords"}

tesei.NewPipeline[files.TextFile]().
    Sequential(files.ListDir{Path: "./data"}).
    Sequential(files.ReadFile{}).
    // Parallel broadcasts the SAME message to all branches.
    // Each branch receives a clone of the message.
    Parallel(summarizeJob, keywordsJob). 
    // Note: You'll likely want a merge step or independent sinks after this
    Build()
```
*Note: `Parallel` broadcasts messages. If you want to split work across workers, use `FanOut`.*

### 3. High-Throughput Worker Pool (Fan-Out)
**Scenario**: You have a large queue of items and want to process them using a pool of workers to maximize throughput.

```go
tesei.NewPipeline[string]().
    Sequential(sourceJob).
    // Spawns 10 workers consuming from the same input channel
    FanOut(heavyComputationJob, 10). 
    Sequential(tesei.End[string]{}).
    Build()
```

### 4. Advanced: Nested Pipelines
**Scenario**: You want to encapsulate a complex sequence of steps into a reusable component.

```go
// Create a sub-pipeline (note: no End job, so it passes data through)
subPipeline := tesei.NewPipeline[string]().
    Sequential(step1).
    Sequential(step2).
    Build()

// Use the sub-pipeline as a job in the main pipeline
mainPipeline := tesei.NewPipeline[string]().
    Sequential(source).
    Sequential(subPipeline). // Executor implements Job!
    Sequential(tesei.End[string]{}).
    Build()
```

## License

MIT License. See [LICENSE](LICENSE) for details.