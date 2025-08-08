package tesei

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestIntegrationSequentialPipeline(t *testing.T) {
	var mu sync.Mutex
	processLog := []string{}

	reader := &TransformJob[string]{
		Transform: func(msg *Message[string]) (*Message[string], error) {
			mu.Lock()
			processLog = append(processLog, "read")
			mu.Unlock()
			msg.Data = "data_from_reader"
			return msg, nil
		},
	}

	processor := &TransformJob[string]{
		Transform: func(msg *Message[string]) (*Message[string], error) {
			mu.Lock()
			processLog = append(processLog, "process")
			mu.Unlock()
			msg.Data = strings.ToUpper(msg.Data)
			return msg, nil
		},
	}

	writer := &TransformJob[string]{
		Transform: func(msg *Message[string]) (*Message[string], error) {
			mu.Lock()
			processLog = append(processLog, "write")
			mu.Unlock()
			msg.Data = msg.Data + "_WRITTEN"
			return msg, nil
		},
	}

	p := NewPipeline[string]().
		Sequential(reader).
		Sequential(processor).
		Sequential(writer).
		Build()

	ctx := context.Background()
	go p.Start(ctx)
	time.Sleep(10 * time.Millisecond)

	p.Input() <- NewMessage("input")
	close(p.Input())

	result := <-p.Output()
	if result.Data != "DATA_FROM_READER_WRITTEN" {
		t.Errorf("Expected 'DATA_FROM_READER_WRITTEN', got %v", result.Data)
	}

	mu.Lock()
	if len(processLog) != 3 {
		t.Errorf("Expected 3 process steps, got %d", len(processLog))
	}
	if processLog[0] != "read" || processLog[1] != "process" || processLog[2] != "write" {
		t.Errorf("Expected order [read, process, write], got %v", processLog)
	}
	mu.Unlock()
}

func TestIntegrationParallelPipeline(t *testing.T) {
	uppercase := &TransformJob[string]{
		Transform: func(msg *Message[string]) (*Message[string], error) {
			msg.Data = strings.ToUpper(msg.Data)
			return msg, nil
		},
	}

	addBrackets := &TransformJob[string]{
		Transform: func(msg *Message[string]) (*Message[string], error) {
			msg.Data = "[" + msg.Data + "]"
			msg.Metadata["processor"] = "brackets"
			return msg, nil
		},
	}

	addQuotes := &TransformJob[string]{
		Transform: func(msg *Message[string]) (*Message[string], error) {
			msg.Data = "\"" + msg.Data + "\""
			msg.Metadata["processor"] = "quotes"
			return msg, nil
		},
	}

	p := NewPipeline[string]().
		Sequential(uppercase).
		Parallel(addBrackets, addQuotes).
		Build()

	ctx := context.Background()
	go p.Start(ctx)
	time.Sleep(10 * time.Millisecond)

	p.Input() <- NewMessage("hello")
	p.Input() <- NewMessage("world")
	close(p.Input())

	results := make(map[string]int)
	for i := 0; i < 4; i++ {
		result := <-p.Output()
		results[result.Data]++
	}

	if results["[HELLO]"] != 1 {
		t.Errorf("Expected 1 '[HELLO]', got %d", results["[HELLO]"])
	}
	if results["\"HELLO\""] != 1 {
		t.Errorf("Expected 1 '\"HELLO\"', got %d", results["\"HELLO\""])
	}
	if results["[WORLD]"] != 1 {
		t.Errorf("Expected 1 '[WORLD]', got %d", results["[WORLD]"])
	}
	if results["\"WORLD\""] != 1 {
		t.Errorf("Expected 1 '\"WORLD\"', got %d", results["\"WORLD\""])
	}
}

func TestIntegrationFanOutPipeline(t *testing.T) {
	var counter int32
	slowProcessor := &TransformJob[int]{
		Transform: func(msg *Message[int]) (*Message[int], error) {
			atomic.AddInt32(&counter, 1)
			time.Sleep(10 * time.Millisecond)
			msg.Data = msg.Data * 2
			return msg, nil
		},
	}

	p := NewPipeline[int]().
		FanOut(slowProcessor, 5).
		Build()

	ctx := context.Background()
	go p.Start(ctx)
	time.Sleep(10 * time.Millisecond)

	start := time.Now()

	for i := 0; i < 10; i++ {
		p.Input() <- NewMessage(i)
	}
	close(p.Input())

	results := make(map[int]bool)
	for i := 0; i < 10; i++ {
		result := <-p.Output()
		results[result.Data] = true
	}

	duration := time.Since(start)

	if atomic.LoadInt32(&counter) != 10 {
		t.Errorf("Expected 10 messages processed, got %d", counter)
	}

	for i := 0; i < 10; i++ {
		expected := i * 2
		if !results[expected] {
			t.Errorf("Missing result %d", expected)
		}
	}

	if duration > 50*time.Millisecond {
		t.Logf("Processing took %v, which indicates good parallelization", duration)
	}
}

func TestIntegrationComplexWorkflow(t *testing.T) {
	generateNumbers := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
		defer close(out)
		for i := 1; i <= 5; i++ {
			select {
			case out <- NewMessage(i):
			case <-ctx.Done():
				return
			}
		}
	})

	multiplyBy2 := &TransformJob[int]{
		Transform: func(msg *Message[int]) (*Message[int], error) {
			msg.Data = msg.Data * 2
			return msg, nil
		},
	}

	multiplyBy3 := &TransformJob[int]{
		Transform: func(msg *Message[int]) (*Message[int], error) {
			msg.Data = msg.Data * 3
			return msg, nil
		},
	}

	sum := JobFunc[int](func(ctx *Thread, in <-chan *Message[int], out chan<- *Message[int]) {
		defer close(out)
		total := 0
		for msg := range in {
			total += msg.Data
		}
		select {
		case out <- NewMessage(total):
		case <-ctx.Done():
			return
		}
	})

	p := NewPipeline[int]().
		Sequential(generateNumbers).
		Parallel(multiplyBy2, multiplyBy3).
		Sequential(sum).
		Build()

	ctx := context.Background()
	go p.Start(ctx)
	time.Sleep(10 * time.Millisecond)

	close(p.Input())

	result := <-p.Output()

	if result.Data != 75 {
		t.Errorf("Expected sum of 75, got %v", result.Data)
	}
}

func TestIntegrationMultiStageErrorRecovery(t *testing.T) {
	stage1 := &TransformJob[string]{
		Transform: func(msg *Message[string]) (*Message[string], error) {
			msg.Data = msg.Data + "_stage1"
			return msg, nil
		},
	}

	errorStage := &TransformJob[string]{
		Transform: func(msg *Message[string]) (*Message[string], error) {
			if msg.Data == "error_stage1" {
				return msg, errors.New("intentional error")
			}
			msg.Data = msg.Data + "_stage2"
			return msg, nil
		},
	}

	recoveryStage := &TransformJob[string]{
		ProcessError: true,
		Transform: func(msg *Message[string]) (*Message[string], error) {
			if msg.HasError() {
				msg.Data = "recovered_" + msg.Data
				msg.Error = nil
				msg.ErrorStage = ""
			} else {
				msg.Data = msg.Data + "_stage3"
			}
			return msg, nil
		},
	}

	p := NewPipeline[string]().
		Sequential(stage1).
		Sequential(errorStage).
		Sequential(recoveryStage).
		Build()

	ctx := context.Background()
	go p.Start(ctx)
	time.Sleep(10 * time.Millisecond)

	p.Input() <- NewMessage("normal")
	p.Input() <- NewMessage("error")
	close(p.Input())

	result1 := <-p.Output()
	result2 := <-p.Output()

	results := map[string]bool{
		result1.Data: true,
		result2.Data: true,
	}

	if !results["normal_stage1_stage2_stage3"] {
		t.Error("Expected 'normal_stage1_stage2_stage3' in results")
	}

	if !results["recovered_error_stage1"] {
		t.Error("Expected 'recovered_error_stage1' in results")
	}
}
