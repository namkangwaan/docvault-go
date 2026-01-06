package stream

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
)

// ChunkedWriter writes data in chunks using a buffer pool
type ChunkedWriter struct {
	file         *os.File
	bufferPool   *BufferPool
	hash         hash.Hash
	bytesWritten int64
}

// NewChunkedWriter creates a new chunked writer
func NewChunkedWriter(filePath string, bufferPool *BufferPool) (*ChunkedWriter, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	return &ChunkedWriter{
		file:       file,
		bufferPool: bufferPool,
		hash:       sha256.New(),
	}, nil
}

// Write implements io.Writer interface
func (cw *ChunkedWriter) Write(p []byte) (n int, err error) {
	n, err = cw.file.Write(p)
	if err != nil {
		return n, err
	}

	// Update hash
	if cw.hash != nil {
		cw.hash.Write(p[:n])
	}

	cw.bytesWritten += int64(n)
	return n, nil
}

// Close closes the file and finalizes the hash
func (cw *ChunkedWriter) Close() error {
	if cw.file != nil {
		return cw.file.Close()
	}
	return nil
}

// BytesWritten returns the number of bytes written so far
func (cw *ChunkedWriter) BytesWritten() int64 {
	return cw.bytesWritten
}

// Checksum returns the SHA-256 checksum of the written data
func (cw *ChunkedWriter) Checksum() string {
	if cw.hash != nil {
		return hex.EncodeToString(cw.hash.Sum(nil))
	}
	return ""
}

// CopyFrom copies data from reader to writer using buffer pool
func (cw *ChunkedWriter) CopyFrom(reader io.Reader) (int64, error) {
	buffer := cw.bufferPool.Get()
	defer cw.bufferPool.Put(buffer)

	return io.CopyBuffer(cw, reader, *buffer)
}
