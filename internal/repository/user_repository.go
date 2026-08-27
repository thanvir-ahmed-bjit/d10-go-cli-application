package repository

import (
	"context"

	"d10-go-cli-application/internal/model"
)

// UserRepository defines the interface for user persistence operations.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	List(ctx context.Context) ([]*model.User, error)
	Find(ctx context.Context, id int) (*model.User, error)
	Update(ctx context.Context, id int, name, email string) error
	Delete(ctx context.Context, id int) error
	Filter(ctx context.Context, term string) ([]*model.User, error)
}
