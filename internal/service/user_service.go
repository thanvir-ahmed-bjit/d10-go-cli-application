package service

import (
	"context"
	"fmt"

	"d10-go-cli-application/internal/logger"
	"d10-go-cli-application/internal/model"
	"d10-go-cli-application/internal/repository"
)

// UserService provides business logic for user operations.
type UserService struct {
	repository repository.UserRepository
	logger     logger.Logger
}

// NewUserService creates a new UserService.
func NewUserService(
	repository repository.UserRepository,
	logger logger.Logger,
) *UserService {
	return &UserService{
		repository: repository,
		logger:     logger,
	}
}

// Create adds a new user.
func (s *UserService) Create(ctx context.Context, user *model.User) error {
	if err := s.repository.Create(ctx, user); err != nil {
		return err
	}

	_ = s.logger.Log(ctx, fmt.Sprintf("Created user: ID=%d, Name=%s, Email=%s", user.ID, user.Name, user.Email))
	return nil
}

// List returns all users.
func (s *UserService) List(ctx context.Context) ([]*model.User, error) {
	return s.repository.List(ctx)
}

// Find returns a user by ID.
func (s *UserService) Find(ctx context.Context, id int) (*model.User, error) {
	return s.repository.Find(ctx, id)
}

// Update modifies an existing user.
func (s *UserService) Update(ctx context.Context, id int, name, email string) error {
	if err := s.repository.Update(ctx, id, name, email); err != nil {
		return err
	}

	_ = s.logger.Log(ctx, fmt.Sprintf("Updated user: ID=%d, Name=%s, Email=%s", id, name, email))
	return nil
}

// Delete removes a user.
func (s *UserService) Delete(ctx context.Context, id int) error {
	if err := s.repository.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.logger.Log(ctx, fmt.Sprintf("Deleted user: ID=%d", id))
	return nil
}

// Filter returns users matching a search term.
func (s *UserService) Filter(ctx context.Context, term string) ([]*model.User, error) {
	return s.repository.Filter(ctx, term)
}
