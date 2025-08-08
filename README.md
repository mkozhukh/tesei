# Tesei

A Go library for building data processing pipelines, born from demand to automate mass file processing through LLM

## Installation

```bash
go get github.com/mkozhukh/tesei
```

## Quick Example

```go
package main

import (
    "context"
    "fmt"
    "github.com/mkozhukh/tesei"
    "github.com/mkozhukh/tesei/files"
)

func main() {
    // Create a pipeline that reads text files and processes them
    pipeline := tesei.NewPipeline[files.TextFile]().
        Sequential(files.ListDir{Path: "./data", Ext: ".txt"}).
        Sequential(files.ReadFile{}).
        Sequential(files.CompleteContent{Prompt: "Summarize this text"}).
        Sequential(files.WriteFile{}).
        Sequential(tesei.End[files.TextFile]{}).
        Build()

    err := pipeline.Start(context.Background())
    if err != nil {
        fmt.Println("Pipeline error:", err)
    }
}
```

## Core Concepts

### Message
The basic unit of data flowing through the pipeline:
- `ID`: Unique identifier for tracking
- `Data`: The actual payload (generic type)
- `Metadata`: Key-value pairs for additional context
- `Error` and `ErrorStage`: Error handling information

### Job
The processing unit that operates on messages. Implement the `Job[T]` interface:
```go
type Job[T any] interface {
    Run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T])
}
```

### Pipeline
Chain jobs together using different execution patterns:
- **Sequential**: Process messages one after another
- **Parallel**: Process messages through multiple jobs simultaneously
- **FanOut**: Distribute work across multiple instances of the same job

## Public API

### Pipeline Building

```go
// Create a new pipeline
pipeline := tesei.NewPipeline[T]()

// Add sequential stages
pipeline.Sequential(job1).Sequential(job2)

// Add parallel branches (messages are duplicated to each branch)
pipeline.Parallel(branchJob1, branchJob2)

// Add fan-out stage (messages are distributed across workers)
pipeline.FanOut(workerJob, workerCount)

// Set buffer size for channels
pipeline.WithBufferSize(10)

// Build the executor
executor := pipeline.Build()

// Start the pipeline
err := executor.Start(context.Background())
```

### Built-in Jobs

#### Core Jobs

- `TransformJob[T]`: Apply a transformation function to each message
```go
tesei.TransformJob[string]{
    Transform: func(msg *tesei.Message[string]) (*tesei.Message[string], error) {
        msg.Data = strings.ToUpper(msg.Data)
        return msg, nil
    },
}
```

- `End[T]`: Terminal job that consumes messages
```go
tesei.End[string]{Log: true} // Optionally log processed messages
```

- `Log[T]`: Pass-through job that logs messages
```go
tesei.Log[string]{Message: "Processing step"}
```

- `SetMetaData[T]`: Add metadata to messages
```go
tesei.SetMetaData[string]{Key: "source", Value: "api"}
```

- `StringsSource`: Generate messages from a slice of strings
```go
tesei.StringsSource{Strings: []string{"one", "two", "three"}}
```

#### File Processing Jobs (files package)

- `ListDir`: List files in a directory
```go
files.ListDir{Path: "./data", Ext: ".txt"}
```

- `ReadFile`: Read file contents
```go
files.ReadFile{}
```

- `WriteFile`: Write file contents
```go
files.WriteFile{DryRun: false}
```

- `PrintContent`: Print file contents to stdout
```go
files.PrintContent{}
```

#### LLM Integration Jobs (files package)

- `CompleteContent`: Process content with LLM
```go
files.CompleteContent{
    Prompt: "Summarize this text",
    Model: "google/fast",
}
```

- `CompleteTemplateString`: Process content with LLM by using inline template
```go
files.CompleteTemplateString{
    Template: "@system: You are a helpful assistant\n@user: {{user_query}}",
}
```

- `CompleteTemplate`: Process content with LLM by using template from file system
```go
files.CompleteTemplate{
    Template: "summarize", // References ~/.prompts/summarize.md
}
```

## Creating Custom Jobs

### Method 1: Use Transform Helper

```go
type MyJob struct{}

func (j MyJob) Run(ctx *tesei.Thread, in <-chan *tesei.Message[string], out chan<- *tesei.Message[string]) {
    tesei.Transform(ctx, in, out, func(msg *tesei.Message[string]) (*tesei.Message[string], error) {
        // Process the message
        msg.Data = processData(msg.Data)
        return msg, nil
    })
}
```

### Method 2: Implement Job Interface

```go
type WordCounter struct {
    MinLength int
}

func (w WordCounter) Run(ctx *tesei.Thread, in <-chan *tesei.Message[string], out chan<- *tesei.Message[string]) {
    defer close(out)
    for {
        select {
        case msg, ok := <-in:
            if !ok {
                return
            }
            words := strings.Fields(msg.Data)
            count := 0
            for _, word := range words {
                if len(word) >= w.MinLength {
                    count++
                }
            }
            msg.Metadata["word_count"] = count
            select {
            case out <- msg:
            case <-ctx.Done():
                return
            }
        case <-ctx.Done():
            return
        }
    }
}
```

- job must close the `out` channel
- job must be ready to exit on `ctx.Done`


## Error Handling

There are two types of errors

### job initialization error

reports that job can't work properly at all, need to be reported through context

```go
ctx.Error() <- err
```

### message proceessing errors

error related to processing specific file, stored on message

```go
if err != nil {
    msg.Error = err
}
```

message with error still need to be pushed in the output channel

such message will be ignored by predefined jobs provided by package ( except of Log/End jobs)

if you are using TransformJob helper you can set `ProcessError` flag to receive errors as well

```go
recoveryJob := tesei.TransformJob[string]{
    ProcessError: true,
    Transform: func(msg *tesei.Message[string]) (*tesei.Message[string], error) {
        if msg.HasError() {
            // Handle the error
            msg.Data = "recovered: " + msg.Data
            msg.Error = nil
            msg.ErrorStage = ""
        }
        return msg, nil
    },
}
```

## Configuration

### LLM Configuration (files package)

```go
// Set global model
files.SetModel("google/gemini-1.5-pro")

// Set API key
files.SetAPIKey("your-api-key")

// Set templates path
files.SetTemplatesPath("~/.prompts")

// Or use custom template source
files.SetTemplatesSource(customSource)
```

## License

MIT License

See LICENSE file for details.