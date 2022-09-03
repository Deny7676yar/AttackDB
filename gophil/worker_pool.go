package main

import (
	"context"
	"sync"
)

type WorkerPool struct {
	workers     chan struct{}
	isClosed    bool
	isClosedMux *sync.Mutex
	ctx         context.Context
	cancelCtx   context.CancelFunc
}

func NewWorkerPool(ctx context.Context, workersCount int) *WorkerPool {
	ctx, cancelCtx := context.WithCancel(ctx)
	workers := make(chan struct{}, workersCount)
	for i := 0; i < workersCount; i++ {
		workers <- struct{}{}
	}
	return &WorkerPool{
		workers:     workers,
		isClosed:    false,
		isClosedMux: &sync.Mutex{},
		ctx:         ctx,
		cancelCtx:   cancelCtx,
	}
}

func (wp *WorkerPool) Run(fn func(ctx context.Context) error) <-chan error {
	result := make(chan error, 1)
	select {
	case _, ok := <-wp.workers:
		if !ok {
			result <- wp.ctx.Err()
			close(result)
			return result
		}
	case <-wp.ctx.Done():
		result <- wp.ctx.Err()
		close(result)
		return result
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		err := fn(wp.ctx)
		result <- err
		close(result)
	}()
	go func() {
		<-done
		wp.addToWorkersPool()
	}()
	return result
}

func (wp *WorkerPool) Close() {
	wp.cancelCtx()
	wp.closeWorkersPool()
}

func (wp *WorkerPool) closeWorkersPool() {
	wp.isClosedMux.Lock()
	defer wp.isClosedMux.Unlock()

	if wp.isClosed {
		return
	}
	close(wp.workers)
}

func (wp *WorkerPool) addToWorkersPool() {
	wp.isClosedMux.Lock()
	defer wp.isClosedMux.Unlock()

	if wp.isClosed {
		return
	}
	wp.workers <- struct{}{}
}
