package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID         int64     `json:"id" db:"id"`
	UserID     *int64    `json:"user_id,omitempty" db:"user_id"`
	Action     string    `json:"action" db:"action"`
	EntityType string    `json:"entity_type" db:"entity_type"`
	EntityID   *int64    `json:"entity_id,omitempty" db:"entity_id"`
	OldValues  JSONB     `json:"old_values,omitempty" db:"old_values"`
	NewValues  JSONB     `json:"new_values,omitempty" db:"new_values"`
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// JSONB is a custom type for PostgreSQL JSONB columns
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, j)
}
