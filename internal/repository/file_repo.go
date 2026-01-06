package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/namkangwaan/docvault-go/internal/models"
)

type FileRepository struct {
	db *pgxpool.Pool
}

func NewFileRepository(db *pgxpool.Pool) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, file *models.File) error {
	query := `
		INSERT INTO files (document_id, file_name, original_name, mime_type, file_extension,
		                   file_size, storage_path, checksum, chunk_size, total_chunks,
		                   file_category, sort_order, preview_generated, preview_path)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		file.DocumentID, file.FileName, file.OriginalName, file.MimeType, file.FileExtension,
		file.FileSize, file.StoragePath, file.Checksum, file.ChunkSize, file.TotalChunks,
		file.FileCategory, file.SortOrder, file.PreviewGenerated, file.PreviewPath,
	).Scan(&file.ID, &file.CreatedAt, &file.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

func (r *FileRepository) GetByID(ctx context.Context, id int64) (*models.File, error) {
	file := &models.File{}
	query := `
		SELECT id, document_id, file_name, original_name, mime_type, file_extension,
		       file_size, storage_path, checksum, chunk_size, total_chunks,
		       file_category, sort_order, preview_generated, preview_path, created_at, updated_at
		FROM files
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&file.ID, &file.DocumentID, &file.FileName, &file.OriginalName, &file.MimeType, &file.FileExtension,
		&file.FileSize, &file.StoragePath, &file.Checksum, &file.ChunkSize, &file.TotalChunks,
		&file.FileCategory, &file.SortOrder, &file.PreviewGenerated, &file.PreviewPath,
		&file.CreatedAt, &file.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return file, nil
}

func (r *FileRepository) ListByDocumentID(ctx context.Context, documentID int64) ([]models.File, error) {
	query := `
		SELECT id, document_id, file_name, original_name, mime_type, file_extension,
		       file_size, storage_path, checksum, chunk_size, total_chunks,
		       file_category, sort_order, preview_generated, preview_path, created_at, updated_at
		FROM files
		WHERE document_id = $1
		ORDER BY sort_order, created_at
	`

	rows, err := r.db.Query(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		err := rows.Scan(
			&file.ID, &file.DocumentID, &file.FileName, &file.OriginalName, &file.MimeType, &file.FileExtension,
			&file.FileSize, &file.StoragePath, &file.Checksum, &file.ChunkSize, &file.TotalChunks,
			&file.FileCategory, &file.SortOrder, &file.PreviewGenerated, &file.PreviewPath,
			&file.CreatedAt, &file.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}
		files = append(files, file)
	}

	return files, nil
}

func (r *FileRepository) Update(ctx context.Context, file *models.File) error {
	query := `
		UPDATE files
		SET file_name = $1, original_name = $2, mime_type = $3, file_extension = $4,
		    file_size = $5, storage_path = $6, checksum = $7, chunk_size = $8, total_chunks = $9,
		    file_category = $10, sort_order = $11, preview_generated = $12, preview_path = $13
		WHERE id = $14
	`

	_, err := r.db.Exec(ctx, query,
		file.FileName, file.OriginalName, file.MimeType, file.FileExtension,
		file.FileSize, file.StoragePath, file.Checksum, file.ChunkSize, file.TotalChunks,
		file.FileCategory, file.SortOrder, file.PreviewGenerated, file.PreviewPath, file.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	return nil
}

func (r *FileRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM files WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (r *FileRepository) DeleteByDocumentID(ctx context.Context, documentID int64) error {
	query := `DELETE FROM files WHERE document_id = $1`

	_, err := r.db.Exec(ctx, query, documentID)
	if err != nil {
		return fmt.Errorf("failed to delete files: %w", err)
	}

	return nil
}
