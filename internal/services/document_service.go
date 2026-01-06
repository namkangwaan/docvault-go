package services

import (
	"context"
	"fmt"

	"github.com/namkangwaan/docvault-go/internal/models"
	"github.com/namkangwaan/docvault-go/internal/repository"
)

type DocumentService struct {
	documentRepo *repository.DocumentRepository
	fileRepo     *repository.FileRepository
}

func NewDocumentService(documentRepo *repository.DocumentRepository, fileRepo *repository.FileRepository) *DocumentService {
	return &DocumentService{
		documentRepo: documentRepo,
		fileRepo:     fileRepo,
	}
}

func (s *DocumentService) Create(ctx context.Context, doc *models.Document) error {
	return s.documentRepo.Create(ctx, doc)
}

func (s *DocumentService) GetByID(ctx context.Context, id int64, includeFiles bool) (*models.Document, error) {
	doc, err := s.documentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if includeFiles {
		files, err := s.fileRepo.ListByDocumentID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get files: %w", err)
		}
		doc.Files = files
	}

	return doc, nil
}

func (s *DocumentService) List(ctx context.Context, filter *repository.DocumentFilter) ([]models.Document, int, error) {
	return s.documentRepo.List(ctx, filter)
}

func (s *DocumentService) Update(ctx context.Context, doc *models.Document) error {
	// Check if document exists
	_, err := s.documentRepo.GetByID(ctx, doc.ID)
	if err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	return s.documentRepo.Update(ctx, doc)
}

func (s *DocumentService) Delete(ctx context.Context, id int64) error {
	// This will cascade delete files due to foreign key constraint
	return s.documentRepo.Delete(ctx, id)
}

func (s *DocumentService) IncrementViewCount(ctx context.Context, id int64) error {
	return s.documentRepo.IncrementViewCount(ctx, id)
}

func (s *DocumentService) IncrementDownloadCount(ctx context.Context, id int64) error {
	return s.documentRepo.IncrementDownloadCount(ctx, id)
}

func (s *DocumentService) GetDocumentFiles(ctx context.Context, documentID int64) ([]models.File, error) {
	return s.fileRepo.ListByDocumentID(ctx, documentID)
}
