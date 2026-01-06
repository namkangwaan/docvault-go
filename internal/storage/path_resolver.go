package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PathResolver resolves storage paths based on configuration
type PathResolver struct {
	basePath       string
	organizeByDate bool
}

// NewPathResolver creates a new path resolver
func NewPathResolver(basePath string, organizeByDate bool) *PathResolver {
	return &PathResolver{
		basePath:       basePath,
		organizeByDate: organizeByDate,
	}
}

// ResolvePath resolves the storage path for a file
func (pr *PathResolver) ResolvePath(fileName string, date time.Time) string {
	if pr.organizeByDate {
		// Organize by year/month/day
		year := date.Format("2006")
		month := date.Format("01")
		day := date.Format("02")
		return filepath.Join(pr.basePath, year, month, day, fileName)
	}
	return filepath.Join(pr.basePath, fileName)
}

// EnsureDirectory ensures that the directory for a path exists
func (pr *PathResolver) EnsureDirectory(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

// GetBasePath returns the base path
func (pr *PathResolver) GetBasePath() string {
	return pr.basePath
}
