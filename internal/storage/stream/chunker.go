package stream

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ChunkUploadSession represents a chunked upload session
type ChunkUploadSession struct {
	ID             string
	FileName       string
	TotalSize      int64
	ChunkSize      int64
	TotalChunks    int
	ReceivedChunks map[int]bool
	TempPath       string
	CreatedAt      time.Time
	LastUpdated    time.Time
	mu             sync.RWMutex
}

// Chunker manages chunked file uploads
type Chunker struct {
	sessions   map[string]*ChunkUploadSession
	tempDir    string
	bufferPool *BufferPool
	mu         sync.RWMutex
}

// NewChunker creates a new chunker
func NewChunker(tempDir string, bufferPool *BufferPool) *Chunker {
	return &Chunker{
		sessions:   make(map[string]*ChunkUploadSession),
		tempDir:    tempDir,
		bufferPool: bufferPool,
	}
}

// InitSession initializes a new chunked upload session
func (c *Chunker) InitSession(sessionID, fileName string, totalSize, chunkSize int64) (*ChunkUploadSession, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	totalChunks := int(totalSize / chunkSize)
	if totalSize%chunkSize != 0 {
		totalChunks++
	}

	tempPath := filepath.Join(c.tempDir, sessionID)
	if err := os.MkdirAll(tempPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	session := &ChunkUploadSession{
		ID:             sessionID,
		FileName:       fileName,
		TotalSize:      totalSize,
		ChunkSize:      chunkSize,
		TotalChunks:    totalChunks,
		ReceivedChunks: make(map[int]bool),
		TempPath:       tempPath,
		CreatedAt:      time.Now(),
		LastUpdated:    time.Now(),
	}

	c.sessions[sessionID] = session
	return session, nil
}

// SaveChunk saves a chunk to the session
func (c *Chunker) SaveChunk(sessionID string, chunkIndex int, data []byte) error {
	c.mu.RLock()
	session, exists := c.sessions[sessionID]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	chunkPath := filepath.Join(session.TempPath, fmt.Sprintf("chunk_%d", chunkIndex))
	if err := os.WriteFile(chunkPath, data, 0600); err != nil {
		return fmt.Errorf("failed to save chunk: %w", err)
	}

	session.ReceivedChunks[chunkIndex] = true
	session.LastUpdated = time.Now()

	return nil
}

// IsComplete checks if all chunks have been received
func (c *Chunker) IsComplete(sessionID string) (bool, error) {
	c.mu.RLock()
	session, exists := c.sessions[sessionID]
	c.mu.RUnlock()

	if !exists {
		return false, fmt.Errorf("session not found: %s", sessionID)
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	return len(session.ReceivedChunks) == session.TotalChunks, nil
}

// Finalize combines all chunks into the final file
func (c *Chunker) Finalize(sessionID, outputPath string) error {
	c.mu.RLock()
	session, exists := c.sessions[sessionID]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	// Check if all chunks are received
	if len(session.ReceivedChunks) != session.TotalChunks {
		return fmt.Errorf("not all chunks received: %d/%d", len(session.ReceivedChunks), session.TotalChunks)
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Combine chunks
	buffer := c.bufferPool.Get()
	defer c.bufferPool.Put(buffer)

	for i := 0; i < session.TotalChunks; i++ {
		chunkPath := filepath.Join(session.TempPath, fmt.Sprintf("chunk_%d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return fmt.Errorf("failed to open chunk %d: %w", i, err)
		}

		if _, err := outFile.ReadFrom(chunkFile); err != nil {
			chunkFile.Close()
			return fmt.Errorf("failed to copy chunk %d: %w", i, err)
		}
		chunkFile.Close()
	}

	return nil
}

// CleanupSession removes the session and its temporary files
func (c *Chunker) CleanupSession(sessionID string) error {
	c.mu.Lock()
	session, exists := c.sessions[sessionID]
	if exists {
		delete(c.sessions, sessionID)
	}
	c.mu.Unlock()

	if !exists {
		return nil
	}

	// Remove temporary directory
	if err := os.RemoveAll(session.TempPath); err != nil {
		return fmt.Errorf("failed to remove temp directory: %w", err)
	}

	return nil
}

// GetSession returns session information
func (c *Chunker) GetSession(sessionID string) (*ChunkUploadSession, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	session, exists := c.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// Progress returns the upload progress for a session
func (c *Chunker) Progress(sessionID string) (int, int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	session, exists := c.sessions[sessionID]
	if !exists {
		return 0, 0, fmt.Errorf("session not found: %s", sessionID)
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	return len(session.ReceivedChunks), session.TotalChunks, nil
}
