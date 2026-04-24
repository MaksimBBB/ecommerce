package repository

import (
	"context"
	database "ecommerce/internal/db"
	models "ecommerce/internal/domain"
	"fmt"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
}

type userRepo struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) UserRepository {
	return &userRepo{db: db}
}

func (u *userRepo) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	user.ID = uuid.New()
	return u.db.QueryRowContext(
		ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.Surname,
		user.Role,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (u *userRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := u.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (u *userRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	err := u.db.GetContext(ctx, &user, query, email)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil

}

func (u *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	err := u.db.GetContext(ctx, &user, query, id)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

func (u *userRepo) List(ctx context.Context, limit int, offset int) ([]*models.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var users []*models.User
	err := u.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

func (u *userRepo) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at
	`

	return u.db.QueryRowContext(
		ctx, query,
		user.FirstName,
		user.Surname,
		user.ID,
	).Scan(&user.UpdatedAt)
}
