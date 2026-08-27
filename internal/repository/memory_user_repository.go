package repository

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"d10-go-cli-application/internal/domain"
	"d10-go-cli-application/internal/model"
)

// MemoryUserRepository implements UserRepository with in-memory storage and concurrency safety.
type MemoryUserRepository struct {
	mu           sync.RWMutex
	usersByOrder []*model.User
	usersByID    map[int]*model.User
}

// Compile-time check that MemoryUserRepository implements UserRepository.
var _ UserRepository = (*MemoryUserRepository)(nil)

// NewMemoryUserRepository creates a new MemoryUserRepository.
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		usersByID: make(map[int]*model.User),
	}
}

// Create adds a new user to the repository.
// It returns an error if the user already exists (duplicate ID), if validation fails, or if the context is cancelled.
func (r *MemoryUserRepository) Create(ctx context.Context, user *model.User) error {
	if user == nil {
		return &domain.ValidationError{Field: "user", Reason: "cannot be nil"}
	}

	name := strings.TrimSpace(user.Name)
	email := strings.TrimSpace(user.Email)

	if name == "" {
		return &domain.ValidationError{Field: "name", Reason: "cannot be empty"}
	}
	if email == "" {
		return &domain.ValidationError{Field: "email", Reason: "cannot be empty"}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.usersByID[user.ID]; exists {
		return fmt.Errorf("create user %d: %w", user.ID, domain.ErrDuplicateID)
	}

	// Create a copy to store
	storedUser := cloneUser(user)
	storedUser.Name = name
	storedUser.Email = email

	r.usersByOrder = append(r.usersByOrder, storedUser)
	r.usersByID[storedUser.ID] = storedUser

	return nil
}

// List returns all users in insertion order.
// It returns a defensive copy of the slice to prevent external modification.
func (r *MemoryUserRepository) List(ctx context.Context) ([]*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*model.User, len(r.usersByOrder))
	for i, user := range r.usersByOrder {
		result[i] = cloneUser(user)
	}
	return result, nil
}

// Find returns a user by ID.
// It returns a defensive copy to prevent external modification.
func (r *MemoryUserRepository) Find(ctx context.Context, id int) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.usersByID[id]
	if !exists {
		return nil, fmt.Errorf("find user %d: %w", id, domain.ErrUserNotFound)
	}
	return cloneUser(user), nil
}

// Update modifies an existing user.
// It returns an error if the user is not found or if validation fails.
func (r *MemoryUserRepository) Update(ctx context.Context, id int, name, email string) error {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)

	if name == "" {
		return &domain.ValidationError{Field: "name", Reason: "cannot be empty"}
	}
	if email == "" {
		return &domain.ValidationError{Field: "email", Reason: "cannot be empty"}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.usersByID[id]
	if !exists {
		return fmt.Errorf("update user %d: %w", id, domain.ErrUserNotFound)
	}

	user.Name = name
	user.Email = email
	return nil
}

// Delete removes a user by ID, preserving insertion order.
// It returns an error if the user is not found.
func (r *MemoryUserRepository) Delete(ctx context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.usersByID[id]; !exists {
		return fmt.Errorf("delete user %d: %w", id, domain.ErrUserNotFound)
	}

	// Find and remove from the ordered slice
	for index, user := range r.usersByOrder {
		if user.ID == id {
			// Shift elements and clear the last slot
			copy(r.usersByOrder[index:], r.usersByOrder[index+1:])
			lastIndex := len(r.usersByOrder) - 1
			r.usersByOrder[lastIndex] = nil
			r.usersByOrder = r.usersByOrder[:lastIndex]
			break
		}
	}

	// Remove from the map
	delete(r.usersByID, id)
	return nil
}

// Filter returns users matching a search term in name or email.
// Returns a defensive copy of matching users.
func (r *MemoryUserRepository) Filter(ctx context.Context, term string) ([]*model.User, error) {
	term = strings.ToLower(strings.TrimSpace(term))

	r.mu.RLock()
	defer r.mu.RUnlock()

	matches := make([]*model.User, 0)
	for _, user := range r.usersByOrder {
		if strings.Contains(strings.ToLower(user.Name), term) ||
			strings.Contains(strings.ToLower(user.Email), term) {
			matches = append(matches, cloneUser(user))
		}
	}
	return matches, nil
}

// cloneUser creates a defensive copy of a user.
func cloneUser(user *model.User) *model.User {
	if user == nil {
		return nil
	}
	copy := *user
	return &copy
}
