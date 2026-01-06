package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/namkangwaan/docvault-go/internal/models"
)

type DocumentRepository struct {
	db *pgxpool.Pool
}

func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{db: db}
}

type DocumentFilter struct {
	CategoryID        *int64
	SearchQuery       string
	OrderNumber       string
	EffectiveDateFrom *time.Time
	EffectiveDateTo   *time.Time
	IsPublished       *bool
	Limit             int
	Offset            int
}

func (r *DocumentRepository) Create(ctx context.Context, doc *models.Document) error {
	query := `
		INSERT INTO documents (category_id, order_number, title, description, effective_date, is_published)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, view_count, download_count, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		doc.CategoryID, doc.OrderNumber, doc.Title, doc.Description, doc.EffectiveDate, doc.IsPublished,
	).Scan(&doc.ID, &doc.ViewCount, &doc.DownloadCount, &doc.CreatedAt, &doc.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

func (r *DocumentRepository) GetByID(ctx context.Context, id int64) (*models.Document, error) {
	doc := &models.Document{}
	query := `
		SELECT d.id, d.category_id, d.order_number, d.title, d.description, d.effective_date,
		       d.is_published, d.view_count, d.download_count, d.created_at, d.updated_at,
		       c.id, c.name, c.slug
		FROM documents d
		LEFT JOIN categories c ON d.category_id = c.id
		WHERE d.id = $1
	`

	doc.Category = &models.Category{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.CategoryID, &doc.OrderNumber, &doc.Title, &doc.Description, &doc.EffectiveDate,
		&doc.IsPublished, &doc.ViewCount, &doc.DownloadCount, &doc.CreatedAt, &doc.UpdatedAt,
		&doc.Category.ID, &doc.Category.Name, &doc.Category.Slug,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return doc, nil
}

func (r *DocumentRepository) List(ctx context.Context, filter *DocumentFilter) ([]models.Document, int, error) {
	// Build query
	var conditions []string
	var args []interface{}
	argCount := 0

	if filter.CategoryID != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("d.category_id = $%d", argCount))
		args = append(args, *filter.CategoryID)
	}

	if filter.SearchQuery != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("(d.title ILIKE $%d OR d.order_number ILIKE $%d)", argCount, argCount))
		args = append(args, "%"+filter.SearchQuery+"%")
	}

	if filter.OrderNumber != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("d.order_number = $%d", argCount))
		args = append(args, filter.OrderNumber)
	}

	if filter.EffectiveDateFrom != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("d.effective_date >= $%d", argCount))
		args = append(args, *filter.EffectiveDateFrom)
	}

	if filter.EffectiveDateTo != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("d.effective_date <= $%d", argCount))
		args = append(args, *filter.EffectiveDateTo)
	}

	if filter.IsPublished != nil {
		argCount++
		conditions = append(conditions, fmt.Sprintf("d.is_published = $%d", argCount))
		args = append(args, *filter.IsPublished)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM documents d" + whereClause
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Get documents
	query := `
		SELECT d.id, d.category_id, d.order_number, d.title, d.description, d.effective_date,
		       d.is_published, d.view_count, d.download_count, d.created_at, d.updated_at,
		       c.id, c.name, c.slug
		FROM documents d
		LEFT JOIN categories c ON d.category_id = c.id
	` + whereClause + `
		ORDER BY d.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", argCount+1) + ` OFFSET $` + fmt.Sprintf("%d", argCount+2)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var documents []models.Document
	for rows.Next() {
		var doc models.Document
		doc.Category = &models.Category{}
		err := rows.Scan(
			&doc.ID, &doc.CategoryID, &doc.OrderNumber, &doc.Title, &doc.Description, &doc.EffectiveDate,
			&doc.IsPublished, &doc.ViewCount, &doc.DownloadCount, &doc.CreatedAt, &doc.UpdatedAt,
			&doc.Category.ID, &doc.Category.Name, &doc.Category.Slug,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, doc)
	}

	return documents, total, nil
}

func (r *DocumentRepository) Update(ctx context.Context, doc *models.Document) error {
	query := `
		UPDATE documents
		SET category_id = $1, order_number = $2, title = $3, description = $4,
		    effective_date = $5, is_published = $6
		WHERE id = $7
	`

	_, err := r.db.Exec(ctx, query,
		doc.CategoryID, doc.OrderNumber, doc.Title, doc.Description,
		doc.EffectiveDate, doc.IsPublished, doc.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

func (r *DocumentRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM documents WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

func (r *DocumentRepository) IncrementViewCount(ctx context.Context, id int64) error {
	query := `UPDATE documents SET view_count = view_count + 1 WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	return nil
}

func (r *DocumentRepository) IncrementDownloadCount(ctx context.Context, id int64) error {
	query := `UPDATE documents SET download_count = download_count + 1 WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment download count: %w", err)
	}

	return nil
}
