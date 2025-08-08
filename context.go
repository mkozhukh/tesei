package tesei

import "context"

type Thread struct {
	context.Context
	errorChan chan error
}

func (t *Thread) SetError(err error) {
	t.errorChan <- err
}

func (t *Thread) Done() <-chan struct{} {
	return t.Context.Done()
}

func (t *Thread) Error() chan error {
	return t.errorChan
}

func (t *Thread) GetError() error {
	select {
	case err := <-t.errorChan:
		return err
	default:
		return nil
	}
}

func NewThread(ctx context.Context, errorBufferSize int) *Thread {
	return &Thread{
		Context:   ctx,
		errorChan: make(chan error, errorBufferSize),
	}
}
