package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/namkangwaan/docvault-go/internal/config"
	"github.com/namkangwaan/docvault-go/internal/storage/stream"
)

// NFSStorage handles file operations on NFS/local filesystem
type NFSStorage struct {
	config       *config.StorageConfig
	bufferPool   *stream.BufferPool
	pathResolver *PathResolver
	chunker      *stream.Chunker
	rangeHandler *stream.RangeHandler
}

// NewNFSStorage creates a new NFS storage manager
func NewNFSStorage(cfg *config.StorageConfig) (*NFSStorage, error) {
	// Create necessary directories
	dirs := []string{
		cfg.BasePath,
		cfg.TempPath,
		cfg.PreviewPath,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	bufferPool := stream.NewBufferPool(cfg.BufferSize)
	pathResolver := NewPathResolver(cfg.BasePath, cfg.OrganizeByDate)
	chunker := stream.NewChunker(cfg.TempPath, bufferPool)
	rangeHandler := stream.NewRangeHandler(bufferPool)

	return &NFSStorage{
		config:       cfg,
		bufferPool:   bufferPool,
		pathResolver: pathResolver,
		chunker:      chunker,
		rangeHandler: rangeHandler,
	}, nil
}

// SaveFile saves a file from reader to storage
func (s *NFSStorage) SaveFile(reader io.Reader, originalFileName string) (string, string, int64, error) {
	// Generate unique filename
	ext := filepath.Ext(originalFileName)
	fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Resolve path
	now := time.Now()
	storagePath := s.pathResolver.ResolvePath(fileName, now)

	// Ensure directory exists
	if err := s.pathResolver.EnsureDirectory(storagePath); err != nil {
		return "", "", 0, err
	}

	// Create chunked writer
	writer, err := stream.NewChunkedWriter(storagePath, s.bufferPool)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to create writer: %w", err)
	}
	defer writer.Close()

	// Copy data
	bytesWritten, err := writer.CopyFrom(reader)
	if err != nil {
		// Clean up on error
		os.Remove(storagePath)
		return "", "", 0, fmt.Errorf("failed to save file: %w", err)
	}

	// Check file size limit
	if s.config.MaxFileSize > 0 && bytesWritten > s.config.MaxFileSize {
		os.Remove(storagePath)
		return "", "", 0, fmt.Errorf("file size exceeds limit: %d > %d", bytesWritten, s.config.MaxFileSize)
	}

	checksum := writer.Checksum()

	return storagePath, checksum, bytesWritten, nil
}

// GetFile opens a file for reading
func (s *NFSStorage) GetFile(storagePath string) (*stream.ChunkedReader, error) {
	return stream.NewChunkedReader(storagePath, s.bufferPool)
}

// DeleteFile removes a file from storage
func (s *NFSStorage) DeleteFile(storagePath string) error {
	if err := os.Remove(storagePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// FileExists checks if a file exists
func (s *NFSStorage) FileExists(storagePath string) bool {
	_, err := os.Stat(storagePath)
	return err == nil
}

// GetFileInfo returns file information
func (s *NFSStorage) GetFileInfo(storagePath string) (int64, time.Time, error) {
	info, err := os.Stat(storagePath)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("failed to stat file: %w", err)
	}
	return info.Size(), info.ModTime(), nil
}

// GetChunker returns the chunker for chunked uploads
func (s *NFSStorage) GetChunker() *stream.Chunker {
	return s.chunker
}

// GetRangeHandler returns the range handler for range downloads
func (s *NFSStorage) GetRangeHandler() *stream.RangeHandler {
	return s.rangeHandler
}

// GetBufferPool returns the buffer pool
func (s *NFSStorage) GetBufferPool() *stream.BufferPool {
	return s.bufferPool
}
