package performance

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrBulkheadFull = errors.New("bulkhead is full")
)

// Bulkhead implements the bulkhead pattern for resource isolation
type Bulkhead struct {
	maxConcurrency int
	semaphore      chan struct{}
	mu             sync.RWMutex
	current        int
}

// NewBulkhead creates a new bulkhead
func NewBulkhead(maxConcurrency int) *Bulkhead {
	return &Bulkhead{
		maxConcurrency: maxConcurrency,
		semaphore:      make(chan struct{}, maxConcurrency),
	}
}

// Execute executes a function with bulkhead protection
func (b *Bulkhead) Execute(ctx context.Context, fn func() error) error {
	// Try to acquire semaphore
	select {
	case b.semaphore <- struct{}{}:
		b.mu.Lock()
		b.current++
		b.mu.Unlock()

		defer func() {
			<-b.semaphore
			b.mu.Lock()
			b.current--
			b.mu.Unlock()
		}()

		return fn()
	case <-ctx.Done():
		return ctx.Err()
	default:
		return ErrBulkheadFull
	}
}

// CurrentConcurrency returns current concurrency level
func (b *Bulkhead) CurrentConcurrency() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.current
}

// MaxConcurrency returns max concurrency
func (b *Bulkhead) MaxConcurrency() int {
	return b.maxConcurrency
}

