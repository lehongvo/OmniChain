package performance

import (
	"context"
	"sync"
)

// WorkerPool manages a pool of workers for concurrent task processing
type WorkerPool struct {
	workers   int
	taskQueue chan func()
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workers:   workers,
		taskQueue: make(chan func(), queueSize),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker is a single worker goroutine
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	for {
		select {
		case task := <-wp.taskQueue:
			if task != nil {
				task()
			}
		case <-wp.ctx.Done():
			return
		}
	}
}

// Submit submits a task to the worker pool
func (wp *WorkerPool) Submit(task func()) error {
	select {
	case wp.taskQueue <- task:
		return nil
	case <-wp.ctx.Done():
		return wp.ctx.Err()
	}
}

// Stop stops the worker pool gracefully
func (wp *WorkerPool) Stop() {
	close(wp.taskQueue)
	wp.cancel()
	wp.wg.Wait()
}

// FanOut distributes tasks to multiple workers
func FanOut(input <-chan interface{}, workers int, fn func(interface{})) {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range input {
				fn(item)
			}
		}()
	}
	wg.Wait()
}

// FanIn merges multiple channels into one
func FanIn(inputs ...<-chan interface{}) <-chan interface{} {
	output := make(chan interface{})
	var wg sync.WaitGroup

	for _, input := range inputs {
		wg.Add(1)
		go func(ch <-chan interface{}) {
			defer wg.Done()
			for item := range ch {
				output <- item
			}
		}(input)
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}

// BoundedChannel creates a bounded channel to prevent memory leaks
func BoundedChannel[T any](capacity int) chan T {
	return make(chan T, capacity)
}
