package tesei

var defaultBufferSize = 1

// Pipeline is a builder for creating data processing pipelines.
// It allows chaining stages like Sequential, Parallel, and FanOut.
type Pipeline[T any] struct {
	stages     []stage[T]
	bufferSize int
}

// ErrorHandler is a function type for handling errors in the pipeline.
type ErrorHandler[T any] func(error, *Message[T])

// NewPipeline creates a new pipeline builder for type T.
func NewPipeline[T any]() *Pipeline[T] {
	return &Pipeline[T]{
		stages:     []stage[T]{},
		bufferSize: defaultBufferSize,
	}
}

// Sequential adds one or more jobs to be executed sequentially.
// Each job reads from the previous stage's output and writes to the next stage's input.
func (p *Pipeline[T]) Sequential(jobs ...Job[T]) *Pipeline[T] {
	for _, job := range jobs {
		p.stages = append(p.stages, &sequentialStage[T]{job: job})
	}
	return p
}

// Parallel adds a stage where input messages are broadcast to multiple jobs running in parallel.
// Each job receives a clone of the input message.
func (p *Pipeline[T]) Parallel(jobs ...Job[T]) *Pipeline[T] {
	p.stages = append(p.stages, &parallelStage[T]{jobs: jobs})
	return p
}

// FanOut adds a stage where a single job is run by multiple workers (competing consumers).
// This is useful for increasing throughput of a slow job.
func (p *Pipeline[T]) FanOut(job Job[T], count int) *Pipeline[T] {
	p.stages = append(p.stages, &fanOutStage[T]{
		job:   job,
		count: count,
	})
	return p
}

// WithBufferSize sets the buffer size for channels between stages.
// Default is 1.
func (p *Pipeline[T]) WithBufferSize(size int) *Pipeline[T] {
	p.bufferSize = size
	return p
}

// Build compiles the pipeline and returns an Executor.
// The Executor can be started to run the pipeline.
func (p *Pipeline[T]) Build() Executor[T] {
	return &executor[T]{
		stages:     p.compileStages(),
		bufferSize: p.bufferSize,
	}
}

func (p *Pipeline[T]) compileStages() []stage[T] {
	compiled := make([]stage[T], len(p.stages))
	copy(compiled, p.stages)
	return compiled
}
