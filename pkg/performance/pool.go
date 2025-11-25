package performance

import (
	"strings"
	"sync"
)

// ObjectPool provides object pooling using sync.Pool
type ObjectPool struct {
	pool *sync.Pool
}

// NewObjectPool creates a new object pool
func NewObjectPool(newFunc func() interface{}) *ObjectPool {
	return &ObjectPool{
		pool: &sync.Pool{
			New: newFunc,
		},
	}
}

// Get gets an object from the pool
func (p *ObjectPool) Get() interface{} {
	return p.pool.Get()
}

// Put returns an object to the pool
func (p *ObjectPool) Put(obj interface{}) {
	p.pool.Put(obj)
}

// BufferPool provides byte buffer pooling
var BufferPool = NewObjectPool(func() interface{} {
	return make([]byte, 0, 1024)
})

// StringBuilderPool provides string builder pooling
var StringBuilderPool = NewObjectPool(func() interface{} {
	return &strings.Builder{}
})

