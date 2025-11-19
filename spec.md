# Tesei Specification

## Overview
Tesei is a Go library for building data processing pipelines, designed for automating mass file processing, particularly with LLM integration. It emphasizes a clean, type-safe API using Go generics and a flexible execution model supporting sequential, parallel, and fan-out patterns.

## Core Abstractions

### Message[T]
The fundamental unit of data.
- **Structure**: Contains `ID` (unique), `Data` (payload of type `T`), `Metadata` (map[string]any), and error state (`Error`, `ErrorStage`).
- **Behavior**: Mutable as it flows through the pipeline. Cloned when branching (Parallel/FanOut) to ensure isolation.

### Job[T]
The processing unit.
- **Interface**: `Run(ctx *Thread, in <-chan *Message[T], out chan<- *Message[T])`.
- **Responsibility**: Read from `in`, process, write to `out`. Must handle context cancellation and close `out` when done.

### Pipeline[T]
The structural definition.
- **Role**: Builder pattern to define the sequence of stages.
- **Storage**: Holds a list of `stage[T]` and configuration (e.g., buffer size).
- **Output**: Builds an `Executor[T]`.

### Executor[T]
The runtime engine.
- **Role**: Materializes the pipeline into running goroutines and channels.
- **Lifecycle**: `Start()` initiates processing, returns execution duration and critical errors.
- **Job Compatibility**: `Executor[T]` implements `Job[T]`, allowing pipelines to be nested as jobs within other pipelines.

### Thread
A wrapper around `context.Context`.
- **Role**: Propagates cancellation and carries critical pipeline errors (`SetError`).

## Public API

### Pipeline Construction
- `NewPipeline[T]()`: Initiates a builder.
- `Sequential(jobs ...Job[T])`: Adds linear processing steps.
- `Parallel(jobs ...Job[T])`: Adds branching steps where input is broadcast to all branches.
- `FanOut(job Job[T], count int)`: Adds a worker pool for a single job type.
- `WithBufferSize(int)`: Configures channel buffer size.
- `Build()`: Compiles the pipeline into an `Executor`.

**Important**: The top-level pipeline MUST end with a `tesei.End[T]` job (or equivalent) to consume all messages. Without it, `Start()` will block indefinitely as the output channel fills up. Nested pipelines (used as jobs) do not strictly require `tesei.End` if they are meant to pass data to the parent pipeline, but if they do include it, they will act as sinks.

### Job Implementation
Users implement the `Job[T]` interface.
- **Helpers**:
    - `TransformJob[T]`: Struct-based helper for simple 1:1 transformations.
    - `Transform[T]`: Function helper to handle loop/select boilerplate.
    - `Filter[T]`: Function helper to filtering messages.

## Inner Architecture

### Execution Model
The pipeline is executed as a chain of stages connected by buffered channels.
1.  **Wiring**: `Executor` creates a series of channels connecting stages.
2.  **Goroutines**: Each stage runs in its own goroutine (or multiple for parallel/fan-out).
3.  **Synchronization**: `sync.WaitGroup` tracks active stages. A `done` channel signals completion.

### Stage Types
- **Sequential**: Direct job execution. 1 input -> 1 job -> 1 output.
- **Parallel**:
    -   **Split**: `oneToMany` routine clones incoming messages to N input channels.
    -   **Process**: N jobs run concurrently.
    -   **Merge**: `manyToOne` routine aggregates results from N output channels to 1 pipeline output.
-   **FanOut**:
    -   **Split**: Single input channel shared by N worker goroutines (competing consumers).
    -   **Process**: N instances of the same job run concurrently.
    -   **Merge**: `manyToOne` aggregates results.

### Data Flow & Concurrency
- **Channel Ownership**: Each stage (or the framework helpers) is responsible for closing its output channel(s) when its input is exhausted.
- **Cloning**: In `Parallel` stages, messages are `Clone()`d before being sent to branches to prevent race conditions on mutable data (Data/Metadata).
- **Context**: `Thread` (Context) is passed to all jobs. Cancellation stops all stages.

### Error Handling
Two distinct error flows:
1.  **Item-Level (Recoverable)**:
    -   Stored in `Message.Error` and `Message.ErrorStage`.
    -   Flows through the pipeline.
    -   Standard jobs (like `TransformJob`) skip messages with errors unless `ProcessError` is true.
    -   Terminal jobs (like `End`) or custom recovery jobs can handle them.
2.  **Pipeline-Level (Critical)**:
    -   Reported via `Thread.SetError(err)`.
    -   Causes `Executor.Start` to return the error and cancel the context, stopping the pipeline.

### Tricky Parts / Implementation Notes
-   **Channel Closing**: The `manyToOne` merger must wait for ALL input channels to close before closing its output. This is handled via `sync.WaitGroup` inside the helper.
-   **FanOut Routing**: Unlike `Parallel`, `FanOut` does *not* clone messages. It relies on channel semantics to distribute work to the first available worker.
-   **Context Safety**: All channel operations (send/receive) are wrapped in `select` blocks with `ctx.Done()` to prevent deadlocks during cancellation.
