package security

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// Quarantine manages quarantined files
type Quarantine struct {
	enabled        bool
	quarantinePath string
}

// NewQuarantine creates a new quarantine manager
func NewQuarantine(enabled bool, quarantinePath string) (*Quarantine, error) {
	if enabled {
		if err := os.MkdirAll(quarantinePath, 0700); err != nil {
			return nil, fmt.Errorf("failed to create quarantine directory: %w", err)
		}
	}

	return &Quarantine{
		enabled:        enabled,
		quarantinePath: quarantinePath,
	}, nil
}

// QuarantineFile moves a file to quarantine
func (q *Quarantine) QuarantineFile(filePath, reason string) (string, error) {
	if !q.enabled {
		return "", nil
	}

	// Generate unique quarantine ID
	quarantineID := uuid.New().String()
	quarantinedPath := filepath.Join(q.quarantinePath, quarantineID)

	// Create quarantine subdirectory
	if err := os.MkdirAll(quarantinedPath, 0700); err != nil {
		return "", fmt.Errorf("failed to create quarantine subdirectory: %w", err)
	}

	// Copy file to quarantine
	destPath := filepath.Join(quarantinedPath, filepath.Base(filePath))
	if err := copyFile(filePath, destPath); err != nil {
		return "", fmt.Errorf("failed to copy file to quarantine: %w", err)
	}

	// Create metadata file
	metadataPath := filepath.Join(quarantinedPath, "metadata.txt")
	metadata := fmt.Sprintf(
		"Original Path: %s\nReason: %s\nQuarantined At: %s\n",
		filePath,
		reason,
		time.Now().Format(time.RFC3339),
	)
	if err := os.WriteFile(metadataPath, []byte(metadata), 0644); err != nil {
		return "", fmt.Errorf("failed to write metadata: %w", err)
	}

	// Remove original file
	if err := os.Remove(filePath); err != nil {
		return "", fmt.Errorf("failed to remove original file: %w", err)
	}

	return quarantineID, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
