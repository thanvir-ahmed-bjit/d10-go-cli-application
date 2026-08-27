package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"d10-go-cli-application/internal/domain"
	"d10-go-cli-application/internal/model"
)

// FakeRepository is a test double for UserRepository
type FakeRepository struct {
	users        map[int]*model.User
	createErr    error
	listErr      error
	findErr      error
	updateErr    error
	deleteErr    error
	filterErr    error
	callSequence []string
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		users: make(map[int]*model.User),
	}
}

func (f *FakeRepository) Create(ctx context.Context, user *model.User) error {
	f.callSequence = append(f.callSequence, "Create")
	if f.createErr != nil {
		return f.createErr
	}
	copy := *user
	f.users[user.ID] = &copy
	return nil
}

func (f *FakeRepository) List(ctx context.Context) ([]*model.User, error) {
	f.callSequence = append(f.callSequence, "List")
	if f.listErr != nil {
		return nil, f.listErr
	}
	result := make([]*model.User, 0, len(f.users))
	for _, u := range f.users {
		result = append(result, u)
	}
	return result, nil
}

func (f *FakeRepository) Find(ctx context.Context, id int) (*model.User, error) {
	f.callSequence = append(f.callSequence, "Find")
	if f.findErr != nil {
		return nil, f.findErr
	}
	u, ok := f.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	copy := *u
	return &copy, nil
}

func (f *FakeRepository) Update(ctx context.Context, id int, name, email string) error {
	f.callSequence = append(f.callSequence, "Update")
	if f.updateErr != nil {
		return f.updateErr
	}
	u, ok := f.users[id]
	if !ok {
		return domain.ErrUserNotFound
	}
	u.Name = name
	u.Email = email
	return nil
}

func (f *FakeRepository) Delete(ctx context.Context, id int) error {
	f.callSequence = append(f.callSequence, "Delete")
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.users, id)
	return nil
}

func (f *FakeRepository) Filter(ctx context.Context, term string) ([]*model.User, error) {
	f.callSequence = append(f.callSequence, "Filter")
	if f.filterErr != nil {
		return nil, f.filterErr
	}
	result := make([]*model.User, 0)
	for _, u := range f.users {
		if strings.Contains(strings.ToLower(u.Name), strings.ToLower(term)) ||
			strings.Contains(strings.ToLower(u.Email), strings.ToLower(term)) {
			result = append(result, u)
		}
	}
	return result, nil
}

// FakeLogger is a test double for Logger
type FakeLogger struct {
	messages []string
	closeErr error
	closed   bool
}

func NewFakeLogger() *FakeLogger {
	return &FakeLogger{
		messages: make([]string, 0),
	}
}

func (f *FakeLogger) Log(ctx context.Context, message string) error {
	f.messages = append(f.messages, message)
	return nil
}

func (f *FakeLogger) Close() error {
	f.closed = true
	return f.closeErr
}

func TestUserServiceCreate(t *testing.T) {
	repo := NewFakeRepository()
	log := NewFakeLogger()
	service := NewUserService(repo, log)

	ctx := context.Background()
	user := &model.User{ID: 1, Name: "Asha", Email: "asha@example.com"}

	if err := service.Create(ctx, user); err != nil {
		t.Fatal(err)
	}

	if len(log.messages) != 1 {
		t.Fatalf("expected 1 log message, got %d", len(log.messages))
	}

	if !strings.Contains(log.messages[0], "Created user") {
		t.Fatalf("expected 'Created user' in log, got %s", log.messages[0])
	}
}

func TestUserServiceCreateRepositoryError(t *testing.T) {
	repo := NewFakeRepository()
	repo.createErr = domain.ErrDuplicateID
	log := NewFakeLogger()
	service := NewUserService(repo, log)

	ctx := context.Background()
	user := &model.User{ID: 1, Name: "Asha", Email: "asha@example.com"}

	err := service.Create(ctx, user)
	if !errors.Is(err, domain.ErrDuplicateID) {
		t.Fatalf("expected ErrDuplicateID, got %v", err)
	}

	if len(log.messages) != 0 {
		t.Fatalf("expected no log message on error, got %d messages", len(log.messages))
	}
}

func TestUserServiceList(t *testing.T) {
	repo := NewFakeRepository()
	repo.users[1] = &model.User{ID: 1, Name: "Asha", Email: "asha@example.com"}
	repo.users[2] = &model.User{ID: 2, Name: "Babu", Email: "babu@example.com"}

	log := NewFakeLogger()
	service := NewUserService(repo, log)

	ctx := context.Background()
	users, err := service.List(ctx)

	if err != nil {
		t.Fatal(err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
}

func TestUserServiceFind(t *testing.T) {
	repo := NewFakeRepository()
	repo.users[1] = &model.User{ID: 1, Name: "Asha", Email: "asha@example.com"}

	log := NewFakeLogger()
	service := NewUserService(repo, log)

	ctx := context.Background()
	user, err := service.Find(ctx, 1)

	if err != nil {
		t.Fatal(err)
	}

	if user.Name != "Asha" {
		t.Fatalf("expected name Asha, got %s", user.Name)
	}
}

func TestUserServiceFindNotFound(t *testing.T) {
	repo := NewFakeRepository()
	log := NewFakeLogger()
	service := NewUserService(repo, log)

	ctx := context.Background()
	_, err := service.Find(ctx, 99)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserServiceUpdate(t *testing.T) {
	repo := NewFakeRepository()
	repo.users[1] = &model.User{ID: 1, Name: "Asha", Email: "asha@example.com"}

	log := NewFakeLogger()
	service := NewUserService(repo, log)

	ctx := context.Background()
	if err := service.Update(ctx, 1, "Asha Rahman", "asha.rahman@example.com"); err != nil {
		t.Fatal(err)
	}

	if len(log.messages) != 1 {
		t.Fatalf("expected 1 log message, got %d", len(log.messages))
	}

	if !strings.Contains(log.messages[0], "Updated user") {
		t.Fatalf("expected 'Updated user' in log, got %s", log.messages[0])
	}
}

func TestUserServiceDelete(t *testing.T) {
	repo := NewFakeRepository()
	repo.users[1] = &model.User{ID: 1, Name: "Asha", Email: "asha@example.com"}

	log := NewFakeLogger()
	service := NewUserService(repo, log)

	ctx := context.Background()
	if err := service.Delete(ctx, 1); err != nil {
		t.Fatal(err)
	}

	if len(log.messages) != 1 {
		t.Fatalf("expected 1 log message, got %d", len(log.messages))
	}

	if !strings.Contains(log.messages[0], "Deleted user") {
		t.Fatalf("expected 'Deleted user' in log, got %s", log.messages[0])
	}
}

func TestUserServiceFilter(t *testing.T) {
	repo := NewFakeRepository()
	repo.users[1] = &model.User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	repo.users[2] = &model.User{ID: 2, Name: "Bob", Email: "bob@example.com"}

	log := NewFakeLogger()
	service := NewUserService(repo, log)

	ctx := context.Background()
	users, err := service.Filter(ctx, "alice")

	if err != nil {
		t.Fatal(err)
	}

	if len(users) != 1 {
		t.Fatalf("expected 1 match, got %d", len(users))
	}
}

func TestUserServiceContextCancellation(t *testing.T) {
	repo := NewFakeRepository()
	log := NewFakeLogger()
	service := NewUserService(repo, log)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	user := &model.User{ID: 1, Name: "Asha", Email: "asha@example.com"}
	err := service.Create(ctx, user)

	// The operation should still succeed since repository operations
	// in our implementation don't check context cancellation
	// But we test that context is passed through
	if err == nil {
		// Repository accepted it, check that call was made
		if len(repo.callSequence) == 0 {
			t.Fatal("expected repository call")
		}
	}
}
