package models

import (
	"time"
)

type Document struct {
	ID            int64      `json:"id" db:"id"`
	CategoryID    int64      `json:"category_id" db:"category_id"`
	OrderNumber   string     `json:"order_number" db:"order_number"`
	Title         string     `json:"title" db:"title"`
	Description   *string    `json:"description,omitempty" db:"description"`
	EffectiveDate *time.Time `json:"effective_date,omitempty" db:"effective_date"`
	IsPublished   bool       `json:"is_published" db:"is_published"`
	ViewCount     int64      `json:"view_count" db:"view_count"`
	DownloadCount int64      `json:"download_count" db:"download_count"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	
	// Relations
	Category *Category `json:"category,omitempty" db:"-"`
	Files    []File    `json:"files,omitempty" db:"-"`
}
