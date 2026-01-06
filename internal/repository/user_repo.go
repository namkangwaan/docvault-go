package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/namkangwaan/docvault-go/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User, password string) error {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO admin_users (username, email, password_hash, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query, user.Username, user.Email, string(hashedPassword), user.IsActive).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password_hash, is_active, last_login_at, created_at, updated_at
		FROM admin_users
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password_hash, is_active, last_login_at, created_at, updated_at
		FROM admin_users
		WHERE username = $1
	`

	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id int64) error {
	query := `
		UPDATE admin_users
		SET last_login_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

func (r *UserRepository) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
