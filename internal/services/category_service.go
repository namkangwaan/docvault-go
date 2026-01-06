package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/namkangwaan/docvault-go/internal/models"
	"github.com/namkangwaan/docvault-go/internal/repository"
)

type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *CategoryService) Create(ctx context.Context, category *models.Category) error {
	// Generate slug from name if not provided
	if category.Slug == "" {
		category.Slug = s.generateSlug(category.Name)
	}

	return s.categoryRepo.Create(ctx, category)
}

func (s *CategoryService) GetByID(ctx context.Context, id int64) (*models.Category, error) {
	return s.categoryRepo.GetByID(ctx, id)
}

func (s *CategoryService) List(ctx context.Context, includeInactive bool) ([]models.Category, error) {
	return s.categoryRepo.List(ctx, includeInactive)
}

func (s *CategoryService) Update(ctx context.Context, category *models.Category) error {
	// Check if category exists
	existing, err := s.categoryRepo.GetByID(ctx, category.ID)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// Update slug if name changed
	if existing.Name != category.Name && category.Slug == "" {
		category.Slug = s.generateSlug(category.Name)
	}

	return s.categoryRepo.Update(ctx, category)
}

func (s *CategoryService) Delete(ctx context.Context, id int64) error {
	return s.categoryRepo.Delete(ctx, id)
}

func (s *CategoryService) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	return s.categoryRepo.GetBySlug(ctx, slug)
}

// generateSlug creates a URL-friendly slug from a string
func (s *CategoryService) generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters (keep only alphanumeric and hyphens)
	var result strings.Builder
	for _, char := range slug {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}

	slug = result.String()

	// Remove consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}
