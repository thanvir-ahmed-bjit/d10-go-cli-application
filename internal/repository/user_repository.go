package repository

import (
	"errors"
	"strings"

	"d5/internal/model"
)

var (
	ErrDuplicateID  = errors.New("a user with that ID already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrEmptyName    = errors.New("name cannot be empty")
	ErrEmptyEmail   = errors.New("email cannot be empty")
)

// UserRepository stores users in insertion order and by ID for fast lookup.
type UserRepository struct {
	usersByOrder []*model.User
	usersByID    map[int]*model.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{usersByID: make(map[int]*model.User)}
}

func (r *UserRepository) Create(user *model.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}
	if strings.TrimSpace(user.Name) == "" {
		return ErrEmptyName
	}
	if strings.TrimSpace(user.Email) == "" {
		return ErrEmptyEmail
	}
	if _, exists := r.usersByID[user.ID]; exists {
		return ErrDuplicateID
	}

	r.usersByOrder = append(r.usersByOrder, user)
	r.usersByID[user.ID] = user
	return nil
}

func (r *UserRepository) List() []*model.User {
	users := make([]*model.User, len(r.usersByOrder))
	copy(users, r.usersByOrder)
	return users
}

func (r *UserRepository) Find(id int) (*model.User, bool) {
	user, exists := r.usersByID[id]
	return user, exists
}

func (r *UserRepository) Update(id int, name string, email string) error {
	user, exists := r.usersByID[id]
	if !exists {
		return ErrUserNotFound
	}
	if strings.TrimSpace(name) == "" {
		return ErrEmptyName
	}
	if strings.TrimSpace(email) == "" {
		return ErrEmptyEmail
	}

	user.Name = strings.TrimSpace(name)
	user.Email = strings.TrimSpace(email)
	return nil
}

func (r *UserRepository) Delete(id int) error {
	if _, exists := r.usersByID[id]; !exists {
		return ErrUserNotFound
	}

	for index, user := range r.usersByOrder {
		if user.ID == id {
			copy(r.usersByOrder[index:], r.usersByOrder[index+1:])
			lastIndex := len(r.usersByOrder) - 1
			r.usersByOrder[lastIndex] = nil
			r.usersByOrder = r.usersByOrder[:lastIndex]
			break
		}
	}
	delete(r.usersByID, id)
	return nil
}

func (r *UserRepository) Filter(term string) []*model.User {
	term = strings.ToLower(strings.TrimSpace(term))
	matches := make([]*model.User, 0)
	for _, user := range r.usersByOrder {
		if strings.Contains(strings.ToLower(user.Name), term) || strings.Contains(strings.ToLower(user.Email), term) {
			matches = append(matches, user)
		}
	}
	return matches
}
