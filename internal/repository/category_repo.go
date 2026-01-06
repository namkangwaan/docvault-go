package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/namkangwaan/docvault-go/internal/models"
)

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) error {
	query := `
		INSERT INTO categories (name, slug, parent_id, sort_order, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		category.Name, category.Slug, category.ParentID, category.SortOrder, category.IsActive,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*models.Category, error) {
	category := &models.Category{}
	query := `
		SELECT id, name, slug, parent_id, sort_order, is_active, created_at, updated_at
		FROM categories
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&category.ID, &category.Name, &category.Slug, &category.ParentID,
		&category.SortOrder, &category.IsActive, &category.CreatedAt, &category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}

func (r *CategoryRepository) List(ctx context.Context, includeInactive bool) ([]models.Category, error) {
	query := `
		SELECT c.id, c.name, c.slug, c.parent_id, c.sort_order, c.is_active, c.created_at, c.updated_at,
		       COALESCE(COUNT(d.id), 0) as document_count
		FROM categories c
		LEFT JOIN documents d ON c.id = d.category_id AND d.is_published = true
	`

	if !includeInactive {
		query += " WHERE c.is_active = true"
	}

	query += `
		GROUP BY c.id, c.name, c.slug, c.parent_id, c.sort_order, c.is_active, c.created_at, c.updated_at
		ORDER BY c.sort_order, c.name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(
			&category.ID, &category.Name, &category.Slug, &category.ParentID,
			&category.SortOrder, &category.IsActive, &category.CreatedAt, &category.UpdatedAt,
			&category.DocumentCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *models.Category) error {
	query := `
		UPDATE categories
		SET name = $1, slug = $2, parent_id = $3, sort_order = $4, is_active = $5
		WHERE id = $6
	`

	_, err := r.db.Exec(ctx, query,
		category.Name, category.Slug, category.ParentID, category.SortOrder, category.IsActive, category.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM categories WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

func (r *CategoryRepository) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	category := &models.Category{}
	query := `
		SELECT id, name, slug, parent_id, sort_order, is_active, created_at, updated_at
		FROM categories
		WHERE slug = $1
	`

	err := r.db.QueryRow(ctx, query, slug).Scan(
		&category.ID, &category.Name, &category.Slug, &category.ParentID,
		&category.SortOrder, &category.IsActive, &category.CreatedAt, &category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get category by slug: %w", err)
	}

	return category, nil
}
