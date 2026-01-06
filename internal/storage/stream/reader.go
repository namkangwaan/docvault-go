package stream

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// ChunkedReader reads a file in chunks using a buffer pool
type ChunkedReader struct {
	file       *os.File
	bufferPool *BufferPool
	checksum   string
	bytesRead  int64
	totalSize  int64
}

// NewChunkedReader creates a new chunked reader
func NewChunkedReader(filePath string, bufferPool *BufferPool) (*ChunkedReader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	return &ChunkedReader{
		file:       file,
		bufferPool: bufferPool,
		totalSize:  stat.Size(),
	}, nil
}

// Read implements io.Reader interface
func (cr *ChunkedReader) Read(p []byte) (n int, err error) {
	n, err = cr.file.Read(p)
	cr.bytesRead += int64(n)
	return n, err
}

// Close closes the file
func (cr *ChunkedReader) Close() error {
	if cr.file != nil {
		return cr.file.Close()
	}
	return nil
}

// BytesRead returns the number of bytes read so far
func (cr *ChunkedReader) BytesRead() int64 {
	return cr.bytesRead
}

// TotalSize returns the total size of the file
func (cr *ChunkedReader) TotalSize() int64 {
	return cr.totalSize
}

// CalculateChecksum calculates SHA-256 checksum of the file
func (cr *ChunkedReader) CalculateChecksum() (string, error) {
	if cr.checksum != "" {
		return cr.checksum, nil
	}

	// Reset file position
	if _, err := cr.file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("failed to seek file: %w", err)
	}

	hash := sha256.New()
	buffer := cr.bufferPool.Get()
	defer cr.bufferPool.Put(buffer)

	if _, err := io.CopyBuffer(hash, cr.file, *buffer); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	cr.checksum = hex.EncodeToString(hash.Sum(nil))

	// Reset file position again
	if _, err := cr.file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("failed to reset file position: %w", err)
	}

	return cr.checksum, nil
}

// Seek implements io.Seeker interface
func (cr *ChunkedReader) Seek(offset int64, whence int) (int64, error) {
	return cr.file.Seek(offset, whence)
}
