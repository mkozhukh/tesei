package tesei

import (
	"testing"
)

func TestNewPipeline(t *testing.T) {
	p := NewPipeline[int]()

	if p == nil {
		t.Error("Expected NewPipeline to return non-nil pipeline")
		return
	}

	if p.stages == nil {
		t.Error("Expected stages to be initialized")
	}

	if len(p.stages) != 0 {
		t.Error("Expected empty stages initially")
	}

	if p.bufferSize != defaultBufferSize {
		t.Errorf("Expected default buffer size %d, got %d", defaultBufferSize, p.bufferSize)
	}
}

func TestPipelineSequential(t *testing.T) {
	p := NewPipeline[int]()

	job1 := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
	})

	job2 := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
	})

	result := p.Sequential(job1, job2)

	if result != p {
		t.Error("Expected Sequential to return the same pipeline for chaining")
	}

	if len(p.stages) != 2 {
		t.Errorf("Expected 2 stages, got %d", len(p.stages))
	}

	for i, stage := range p.stages {
		if _, ok := stage.(*sequentialStage[int]); !ok {
			t.Errorf("Expected stage %d to be sequentialStage", i)
		}
	}
}

func TestPipelineParallel(t *testing.T) {
	p := NewPipeline[int]()

	job1 := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
	})

	job2 := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
	})

	result := p.Parallel(job1, job2)

	if result != p {
		t.Error("Expected Parallel to return the same pipeline for chaining")
	}

	if len(p.stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(p.stages))
	}

	parallelStg, ok := p.stages[0].(*parallelStage[int])
	if !ok {
		t.Error("Expected stage to be parallelStage")
	}

	if len(parallelStg.jobs) != 2 {
		t.Errorf("Expected 2 jobs in parallel stage, got %d", len(parallelStg.jobs))
	}
}

func TestPipelineFanOut(t *testing.T) {
	p := NewPipeline[int]()

	job := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
	})

	result := p.FanOut(job, 5)

	if result != p {
		t.Error("Expected FanOut to return the same pipeline for chaining")
	}

	if len(p.stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(p.stages))
	}

	fanOutStg, ok := p.stages[0].(*fanOutStage[int])
	if !ok {
		t.Error("Expected stage to be fanOutStage")
	}

	if fanOutStg.count != 5 {
		t.Errorf("Expected count to be 5, got %d", fanOutStg.count)
	}
}

func TestPipelineWithBufferSize(t *testing.T) {
	p := NewPipeline[int]()

	result := p.WithBufferSize(100)

	if result != p {
		t.Error("Expected WithBufferSize to return the same pipeline for chaining")
	}

	if p.bufferSize != 100 {
		t.Errorf("Expected buffer size to be 100, got %d", p.bufferSize)
	}
}

func TestPipelineChaining(t *testing.T) {
	job1 := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
	})

	job2 := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
	})

	job3 := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
	})

	p := NewPipeline[int]().
		Sequential(job1).
		Parallel(job2, job3).
		FanOut(job1, 3).
		Sequential(job2).
		WithBufferSize(50)

	if len(p.stages) != 4 {
		t.Errorf("Expected 4 stages, got %d", len(p.stages))
	}

	if p.bufferSize != 50 {
		t.Errorf("Expected buffer size 50, got %d", p.bufferSize)
	}

	if _, ok := p.stages[0].(*sequentialStage[int]); !ok {
		t.Error("Expected first stage to be sequentialStage")
	}

	if _, ok := p.stages[1].(*parallelStage[int]); !ok {
		t.Error("Expected second stage to be parallelStage")
	}

	if _, ok := p.stages[2].(*fanOutStage[int]); !ok {
		t.Error("Expected third stage to be fanOutStage")
	}

	if _, ok := p.stages[3].(*sequentialStage[int]); !ok {
		t.Error("Expected fourth stage to be sequentialStage")
	}
}

func TestPipelineBuild(t *testing.T) {
	p := NewPipeline[int]().
		Sequential(JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
			return
		})).
		WithBufferSize(10)

	exec := p.Build()

	if exec == nil {
		t.Error("Expected Build to return non-nil executor")
	}

	execImpl, ok := exec.(*executor[int])
	if !ok {
		t.Error("Expected executor to be of type *executor[int]")
	}

	if len(execImpl.stages) != 1 {
		t.Errorf("Expected 1 stage in executor, got %d", len(execImpl.stages))
	}

	if execImpl.bufferSize != 10 {
		t.Errorf("Expected buffer size 10 in executor, got %d", execImpl.bufferSize)
	}
}

func TestPipelineCompileStages(t *testing.T) {
	p := NewPipeline[int]()

	job := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
		return
	})

	p.Sequential(job, job).Parallel(job, job)

	compiled := p.compileStages()

	if len(compiled) != len(p.stages) {
		t.Errorf("Expected compiled stages to have same length as original, got %d vs %d",
			len(compiled), len(p.stages))
	}

	for i := range compiled {
		if compiled[i] != p.stages[i] {
			t.Errorf("Expected compiled stage %d to match original", i)
		}
	}
}

func TestPipelineNested(t *testing.T) {
	job := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
	})

	processA := NewPipeline[int]().Sequential(job)
	processB := NewPipeline[int]().Sequential(job)

	p := NewPipeline[int]().
		Sequential(job, job).
		Parallel(job, processA.Build()).
		Sequential(processB.Build())

	compiled := p.compileStages()

	if len(compiled) != 4 {
		t.Errorf("Expected 4 stages, got %d", len(compiled))
	}
}
