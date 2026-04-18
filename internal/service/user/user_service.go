package service

import (
	"context"
	models "ecommerce/internal/domain"
	repository "ecommerce/internal/repository/postgres"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetProfile(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, req UpdateProfileRequest, is_admin bool) (*models.User, error)
	List(ctx context.Context, filter UserFilter) ([]*UserListResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	userRepo repository.UserRepository
}

func NewService(userRepo repository.UserRepository) UserService {
	return &service{userRepo: userRepo}
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.userRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete user:%w", err)
	}

	return nil
}

// GetProfile implements [UserService].
func (s *service) GetProfile(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// List implements UserService.
func (s *service) List(ctx context.Context, filter UserFilter) ([]*UserListResponse, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	if filter.Limit > 100 {
		filter.Limit = 100
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}

	users, err := s.userRepo.List(ctx, filter.Limit, filter.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	filtered := []*models.User{}

	for _, u := range users {
		if filter.Role != nil && u.Role != *filter.Role {
			continue
		}

		if filter.Search != "" && !strings.Contains(strings.ToLower(u.FirstName),
			strings.ToLower(filter.Search)) && !strings.Contains(strings.ToLower(u.Surname), strings.ToLower(filter.Search)) {
			continue
		}

		filtered = append(filtered, u)
	}

	resp := &UserListResponse{
		Users:  filtered,
		Total:  len(filtered),
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}

	return []*UserListResponse{resp}, nil
}

// UpdateProfile implements UserService.
func (s *service) UpdateProfile(ctx context.Context, id uuid.UUID, req UpdateProfileRequest, is_admin bool) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	_ = is_admin

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}

	if req.LastName != nil {
		user.Surname = *req.LastName
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	if req.Password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = string(hash)
	}

	if req.Role != nil && is_admin {
		user.Role = *req.Role
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
