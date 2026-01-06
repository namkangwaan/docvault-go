package stream

import (
	"sync"
)

// BufferPool manages a pool of reusable byte buffers to reduce GC pressure
type BufferPool struct {
	pool       *sync.Pool
	bufferSize int
}

// NewBufferPool creates a new buffer pool with the specified buffer size
func NewBufferPool(bufferSize int) *BufferPool {
	return &BufferPool{
		pool: &sync.Pool{
			New: func() interface{} {
				buffer := make([]byte, bufferSize)
				return &buffer
			},
		},
		bufferSize: bufferSize,
	}
}

// Get retrieves a buffer from the pool
func (bp *BufferPool) Get() *[]byte {
	return bp.pool.Get().(*[]byte)
}

// Put returns a buffer to the pool
func (bp *BufferPool) Put(buffer *[]byte) {
	if buffer != nil && len(*buffer) == bp.bufferSize {
		bp.pool.Put(buffer)
	}
}

// Size returns the configured buffer size
func (bp *BufferPool) Size() int {
	return bp.bufferSize
}
