package tesei

import "context"

// Thread is a wrapper around context.Context that also carries pipeline errors.
// It allows propagating critical errors from any stage to the executor.
type Thread struct {
	context.Context
	errorChan chan error
}

// SetError reports a critical error that should stop the pipeline.
func (t *Thread) SetError(err error) {
	t.errorChan <- err
}

// Done returns a channel that's closed when the thread is cancelled.
func (t *Thread) Done() <-chan struct{} {
	return t.Context.Done()
}

// Error returns the channel for reporting errors.
func (t *Thread) Error() chan error {
	return t.errorChan
}

// GetError returns the first error reported to the thread, or nil if none.
func (t *Thread) GetError() error {
	select {
	case err := <-t.errorChan:
		return err
	default:
		return nil
	}
}

// NewThread creates a new Thread with the given context and error buffer size.
func NewThread(ctx context.Context, errorBufferSize int) *Thread {
	return &Thread{
		Context:   ctx,
		errorChan: make(chan error, errorBufferSize),
	}
}
