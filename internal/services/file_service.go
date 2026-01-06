package services

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/namkangwaan/docvault-go/internal/models"
	"github.com/namkangwaan/docvault-go/internal/repository"
	"github.com/namkangwaan/docvault-go/internal/security"
	"github.com/namkangwaan/docvault-go/internal/storage"
)

type FileService struct {
	fileRepo  *repository.FileRepository
	storage   *storage.NFSStorage
	validator *security.Validator
}

func NewFileService(
	fileRepo *repository.FileRepository,
	storage *storage.NFSStorage,
	validator *security.Validator,
) *FileService {
	return &FileService{
		fileRepo:  fileRepo,
		storage:   storage,
		validator: validator,
	}
}

// UploadFile uploads a file with security validation
func (s *FileService) UploadFile(ctx context.Context, reader io.Reader, fileName string, documentID int64) (*models.File, *security.Report, error) {
	// Validate file
	validatedReader, report, err := s.validator.ValidateUpload(reader, fileName)
	if err != nil {
		return nil, report, fmt.Errorf("file validation failed: %w", err)
	}
	defer validatedReader.(io.ReadCloser).Close()

	// Save file to storage
	storagePath, checksum, fileSize, err := s.storage.SaveFile(validatedReader, fileName)
	if err != nil {
		return nil, report, fmt.Errorf("failed to save file: %w", err)
	}

	// Detect MIME type from file extension
	ext := strings.TrimPrefix(filepath.Ext(fileName), ".")
	mimeType := getMimeType(ext)

	// Create file record
	file := &models.File{
		DocumentID:    documentID,
		FileName:      filepath.Base(storagePath),
		OriginalName:  fileName,
		MimeType:      mimeType,
		FileExtension: ext,
		FileSize:      fileSize,
		StoragePath:   storagePath,
		Checksum:      checksum,
		FileCategory:  "attachment",
		SortOrder:     0,
	}

	if err := s.fileRepo.Create(ctx, file); err != nil {
		// Clean up storage on database error
		s.storage.DeleteFile(storagePath)
		return nil, report, fmt.Errorf("failed to create file record: %w", err)
	}

	return file, report, nil
}

// GetFile retrieves a file by ID
func (s *FileService) GetFile(ctx context.Context, id int64) (*models.File, error) {
	return s.fileRepo.GetByID(ctx, id)
}

// GetFileReader opens a file for reading
func (s *FileService) GetFileReader(ctx context.Context, fileID int64) (io.ReadSeekCloser, *models.File, error) {
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, nil, fmt.Errorf("file not found: %w", err)
	}

	reader, err := s.storage.GetFile(file.StoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	return reader, file, nil
}

// DeleteFile deletes a file
func (s *FileService) DeleteFile(ctx context.Context, id int64) error {
	file, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// Delete from storage
	if err := s.storage.DeleteFile(file.StoragePath); err != nil {
		// Log error but continue with database deletion
	}

	// Delete from database
	return s.fileRepo.Delete(ctx, id)
}

// ListFilesByDocument lists all files for a document
func (s *FileService) ListFilesByDocument(ctx context.Context, documentID int64) ([]models.File, error) {
	return s.fileRepo.ListByDocumentID(ctx, documentID)
}

// getMimeType returns MIME type for common file extensions
func getMimeType(ext string) string {
	mimeTypes := map[string]string{
		"pdf":  "application/pdf",
		"docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"doc":  "application/msword",
		"xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"xls":  "application/vnd.ms-excel",
		"pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"ppt":  "application/vnd.ms-powerpoint",
	}

	if mime, ok := mimeTypes[strings.ToLower(ext)]; ok {
		return mime
	}

	return "application/octet-stream"
}

// ReadSeekCloser interface for file readers
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}
