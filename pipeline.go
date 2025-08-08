package tesei

var defaultBufferSize = 1

type Pipeline[T any] struct {
	stages     []stage[T]
	bufferSize int
}

type ErrorHandler[T any] func(error, *Message[T])

func NewPipeline[T any]() *Pipeline[T] {
	return &Pipeline[T]{
		stages:     []stage[T]{},
		bufferSize: defaultBufferSize,
	}
}

func (p *Pipeline[T]) Sequential(jobs ...Job[T]) *Pipeline[T] {
	for _, job := range jobs {
		p.stages = append(p.stages, &sequentialStage[T]{job: job})
	}
	return p
}

func (p *Pipeline[T]) Parallel(jobs ...Job[T]) *Pipeline[T] {
	p.stages = append(p.stages, &parallelStage[T]{jobs: jobs})
	return p
}

func (p *Pipeline[T]) FanOut(job Job[T], count int) *Pipeline[T] {
	p.stages = append(p.stages, &fanOutStage[T]{
		job:   job,
		count: count,
	})
	return p
}

func (p *Pipeline[T]) WithBufferSize(size int) *Pipeline[T] {
	p.bufferSize = size
	return p
}

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
