package stream

import (
	"testing"
)

func TestBufferPool(t *testing.T) {
	bufferSize := 1024
	pool := NewBufferPool(bufferSize)

	// Get a buffer
	buffer := pool.Get()
	if buffer == nil {
		t.Fatal("Expected non-nil buffer")
	}

	if len(*buffer) != bufferSize {
		t.Errorf("Expected buffer size %d, got %d", bufferSize, len(*buffer))
	}

	// Put it back
	pool.Put(buffer)

	// Get another buffer (should be the same one from pool)
	buffer2 := pool.Get()
	if buffer2 == nil {
		t.Fatal("Expected non-nil buffer")
	}

	if len(*buffer2) != bufferSize {
		t.Errorf("Expected buffer size %d, got %d", bufferSize, len(*buffer2))
	}
}

func TestBufferPoolSize(t *testing.T) {
	bufferSize := 2048
	pool := NewBufferPool(bufferSize)

	if pool.Size() != bufferSize {
		t.Errorf("Expected pool size %d, got %d", bufferSize, pool.Size())
	}
}
