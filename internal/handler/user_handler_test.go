package handler

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"d10-go-cli-application/internal/domain"
	"d10-go-cli-application/internal/model"
)

// FakeLogger is a test double for Logger.
type FakeLogger struct {
	messages []string
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

func (f *FakeLogger) Flush() error {
	return nil
}

func (f *FakeLogger) Close() error {
	return nil
}

// FakeService is a test double for UserUseCases.
type FakeService struct {
	users     map[int]*model.User
	createErr error
	listErr   error
	findErr   error
	updateErr error
	deleteErr error
	filterErr error
	callLog   []string
}

func NewFakeService() *FakeService {
	return &FakeService{
		users:   make(map[int]*model.User),
		callLog: make([]string, 0),
	}
}

func (f *FakeService) Create(ctx context.Context, user *model.User) error {
	f.callLog = append(f.callLog, "Create")
	if f.createErr != nil {
		return f.createErr
	}
	copy := *user
	f.users[user.ID] = &copy
	return nil
}

func (f *FakeService) List(ctx context.Context) ([]*model.User, error) {
	f.callLog = append(f.callLog, "List")
	if f.listErr != nil {
		return nil, f.listErr
	}
	result := make([]*model.User, 0)
	for _, u := range f.users {
		result = append(result, u)
	}
	return result, nil
}

func (f *FakeService) Find(ctx context.Context, id int) (*model.User, error) {
	f.callLog = append(f.callLog, "Find")
	if f.findErr != nil {
		return nil, f.findErr
	}
	u, ok := f.users[id]
	if !ok {
		return nil, errors.New("find user 99: " + domain.ErrUserNotFound.Error())
	}
	copy := *u
	return &copy, nil
}

func (f *FakeService) Update(ctx context.Context, id int, name, email string) error {
	f.callLog = append(f.callLog, "Update")
	if f.updateErr != nil {
		return f.updateErr
	}
	u, ok := f.users[id]
	if !ok {
		return errors.New("update user " + string(rune(id)) + ": " + domain.ErrUserNotFound.Error())
	}
	u.Name = name
	u.Email = email
	return nil
}

func (f *FakeService) Delete(ctx context.Context, id int) error {
	f.callLog = append(f.callLog, "Delete")
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.users, id)
	return nil
}

func (f *FakeService) Filter(ctx context.Context, term string) ([]*model.User, error) {
	f.callLog = append(f.callLog, "Filter")
	if f.filterErr != nil {
		return nil, f.filterErr
	}
	result := make([]*model.User, 0)
	term = strings.ToLower(term)
	for _, u := range f.users {
		if strings.Contains(strings.ToLower(u.Name), term) ||
			strings.Contains(strings.ToLower(u.Email), term) {
			result = append(result, u)
		}
	}
	return result, nil
}

func TestUserHandlerAddUser(t *testing.T) {
	service := NewFakeService()
	output := &bytes.Buffer{}
	input := strings.NewReader("1\n1\nAlice\nalice@example.com\n7\n")
	logger := NewFakeLogger()
	handler := NewUserHandler(service, input, output, logger)

	ctx := context.Background()
	handler.Run(ctx)

	if len(service.users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(service.users))
	}

	if !strings.Contains(output.String(), "User added successfully") {
		t.Fatalf("expected success message, got %s", output.String())
	}
}

func TestUserHandlerListUsers(t *testing.T) {
	service := NewFakeService()
	service.users[1] = &model.User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	service.users[2] = &model.User{ID: 2, Name: "Bob", Email: "bob@example.com"}

	output := &bytes.Buffer{}
	input := strings.NewReader("2\n7\n")
	logger := NewFakeLogger()
	handler := NewUserHandler(service, input, output, logger)

	ctx := context.Background()
	handler.Run(ctx)

	result := output.String()
	if !strings.Contains(result, "Alice") || !strings.Contains(result, "Bob") {
		t.Fatalf("expected users in output, got %s", result)
	}
}

func TestUserHandlerFindUser(t *testing.T) {
	service := NewFakeService()
	service.users[1] = &model.User{ID: 1, Name: "Alice", Email: "alice@example.com"}

	output := &bytes.Buffer{}
	input := strings.NewReader("3\n1\n7\n")
	logger := NewFakeLogger()
	handler := NewUserHandler(service, input, output, logger)

	ctx := context.Background()
	handler.Run(ctx)

	result := output.String()
	if !strings.Contains(result, "Alice") {
		t.Fatalf("expected Alice in output, got %s", result)
	}
}

func TestUserHandlerFindUserNotFound(t *testing.T) {
	service := NewFakeService()
	service.findErr = errors.New("find user 99: " + domain.ErrUserNotFound.Error())

	output := &bytes.Buffer{}
	input := strings.NewReader("3\n99\n7\n")
	logger := NewFakeLogger()
	handler := NewUserHandler(service, input, output, logger)

	ctx := context.Background()
	handler.Run(ctx)

	result := output.String()
	if !strings.Contains(result, "Error") {
		t.Fatalf("expected error message, got %s", result)
	}
}

func TestUserHandlerUpdateUser(t *testing.T) {
	service := NewFakeService()
	service.users[1] = &model.User{ID: 1, Name: "Alice", Email: "alice@example.com"}

	output := &bytes.Buffer{}
	input := strings.NewReader("4\n1\nAlice Updated\nalice.updated@example.com\n7\n")
	logger := NewFakeLogger()
	handler := NewUserHandler(service, input, output, logger)

	ctx := context.Background()
	handler.Run(ctx)

	result := output.String()
	if !strings.Contains(result, "User updated successfully") {
		t.Fatalf("expected update success, got %s", result)
	}
}

func TestUserHandlerDeleteUser(t *testing.T) {
	service := NewFakeService()
	service.users[1] = &model.User{ID: 1, Name: "Alice", Email: "alice@example.com"}

	output := &bytes.Buffer{}
	input := strings.NewReader("5\n1\n7\n")
	logger := NewFakeLogger()
	handler := NewUserHandler(service, input, output, logger)

	ctx := context.Background()
	handler.Run(ctx)

	result := output.String()
	if !strings.Contains(result, "User deleted successfully") {
		t.Fatalf("expected delete success, got %s", result)
	}

	if len(service.users) != 0 {
		t.Fatalf("expected 0 users after deletion, got %d", len(service.users))
	}
}

func TestUserHandlerFilterUsers(t *testing.T) {
	service := NewFakeService()
	service.users[1] = &model.User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	service.users[2] = &model.User{ID: 2, Name: "Bob", Email: "bob@alice.com"}

	output := &bytes.Buffer{}
	input := strings.NewReader("6\nalice\n7\n")
	logger := NewFakeLogger()
	handler := NewUserHandler(service, input, output, logger)

	ctx := context.Background()
	handler.Run(ctx)

	result := output.String()
	if !strings.Contains(result, "Alice") {
		t.Fatalf("expected Alice in filter results, got %s", result)
	}
}

func TestUserHandlerInvalidMenuOption(t *testing.T) {
	service := NewFakeService()
	output := &bytes.Buffer{}
	input := strings.NewReader("9\n7\n")
	logger := NewFakeLogger()
	handler := NewUserHandler(service, input, output, logger)

	ctx := context.Background()
	handler.Run(ctx)

	result := output.String()
	if !strings.Contains(result, "Invalid menu option") {
		t.Fatalf("expected invalid option message, got %s", result)
	}
}

func TestUserHandlerInvalidIntegerInput(t *testing.T) {
	service := NewFakeService()
	output := &bytes.Buffer{}
	input := strings.NewReader("3\nabc\n7\n")
	logger := NewFakeLogger()
	handler := NewUserHandler(service, input, output, logger)

	ctx := context.Background()
	handler.Run(ctx)

	result := output.String()
	if !strings.Contains(result, "Invalid integer input") {
		t.Fatalf("expected invalid integer message, got %s", result)
	}
}

func TestUserHandlerEOF(t *testing.T) {
	service := NewFakeService()
	output := &bytes.Buffer{}
	input := strings.NewReader("")
	logger := NewFakeLogger()
	handler := NewUserHandler(service, input, output, logger)

	ctx := context.Background()
	handler.Run(ctx)

	result := output.String()
	if !strings.Contains(result, "Goodbye") {
		t.Fatalf("expected goodbye message on EOF, got %s", result)
	}
}

func TestUserHandlerExit(t *testing.T) {
	service := NewFakeService()
	output := &bytes.Buffer{}
	input := strings.NewReader("7\n")
	logger := NewFakeLogger()
	handler := NewUserHandler(service, input, output, logger)

	ctx := context.Background()
	handler.Run(ctx)

	result := output.String()
	if !strings.Contains(result, "Goodbye") {
		t.Fatalf("expected goodbye message on exit, got %s", result)
	}
}

func TestUserHandlerDuplicateIDError(t *testing.T) {
	service := NewFakeService()
	service.createErr = errors.New("create user 1: " + domain.ErrDuplicateID.Error())

	output := &bytes.Buffer{}
	input := strings.NewReader("1\n1\nAlice\nalice@example.com\n7\n")
	logger := NewFakeLogger()
	handler := NewUserHandler(service, input, output, logger)

	ctx := context.Background()
	handler.Run(ctx)

	result := output.String()
	if !strings.Contains(result, "Error") {
		t.Fatalf("expected error for duplicate ID, got %s", result)
	}
}

func TestUserHandlerValidationErrorRendering(t *testing.T) {
	output := &bytes.Buffer{}

	handler := &UserHandler{output: output}

	// Test ValidationError rendering
	ve := &domain.ValidationError{Field: "name", Reason: "cannot be empty"}
	handler.renderError(ve)

	result := output.String()
	if !strings.Contains(result, "Invalid name") || !strings.Contains(result, "cannot be empty") {
		t.Fatalf("expected validation error format 'Invalid name: cannot be empty', got %q", result)
	}
}
