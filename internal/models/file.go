package models

import (
	"time"
)

type File struct {
	ID               int64     `json:"id" db:"id"`
	DocumentID       int64     `json:"document_id" db:"document_id"`
	FileName         string    `json:"file_name" db:"file_name"`
	OriginalName     string    `json:"original_name" db:"original_name"`
	MimeType         string    `json:"mime_type" db:"mime_type"`
	FileExtension    string    `json:"file_extension" db:"file_extension"`
	FileSize         int64     `json:"file_size" db:"file_size"`
	StoragePath      string    `json:"storage_path" db:"storage_path"`
	Checksum         string    `json:"checksum" db:"checksum"`
	ChunkSize        *int64    `json:"chunk_size,omitempty" db:"chunk_size"`
	TotalChunks      *int      `json:"total_chunks,omitempty" db:"total_chunks"`
	FileCategory     string    `json:"file_category" db:"file_category"`
	SortOrder        int       `json:"sort_order" db:"sort_order"`
	PreviewGenerated bool      `json:"preview_generated" db:"preview_generated"`
	PreviewPath      *string   `json:"preview_path,omitempty" db:"preview_path"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}
