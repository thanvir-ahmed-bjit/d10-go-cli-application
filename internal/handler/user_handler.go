package handler

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"d10-go-cli-application/internal/config"
	"d10-go-cli-application/internal/domain"
	"d10-go-cli-application/internal/model"
)

// UserUseCases defines the use cases the handler consumes.
type UserUseCases interface {
	Create(ctx context.Context, user *model.User) error
	List(ctx context.Context) ([]*model.User, error)
	Find(ctx context.Context, id int) (*model.User, error)
	Update(ctx context.Context, id int, name, email string) error
	Delete(ctx context.Context, id int) error
	Filter(ctx context.Context, term string) ([]*model.User, error)
}

// UserHandler handles CLI interaction for user management.
type UserHandler struct {
	service UserUseCases
	input   *bufio.Scanner
	output  io.Writer
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(service UserUseCases, input io.Reader, output io.Writer) *UserHandler {
	return &UserHandler{
		service: service,
		input:   bufio.NewScanner(input),
		output:  output,
	}
}

// Run starts the interactive CLI loop.
func (h *UserHandler) Run(ctx context.Context) {
	for {
		fmt.Fprintln(h.output, config.MenuHeader)
		fmt.Fprintln(h.output, "1. Add user")
		fmt.Fprintln(h.output, "2. List users")
		fmt.Fprintln(h.output, "3. Find user by ID")
		fmt.Fprintln(h.output, "4. Update user")
		fmt.Fprintln(h.output, "5. Delete user")
		fmt.Fprintln(h.output, "6. Filter users")
		fmt.Fprintln(h.output, "7. Exit")
		choice, ok := h.readLine("Choose an option: ")
		if !ok {
			fmt.Fprintln(h.output, "Goodbye!")
			return
		}

		switch choice {
		case "1":
			h.addUser(ctx)
		case "2":
			h.listUsers(ctx)
		case "3":
			h.findUser(ctx)
		case "4":
			h.updateUser(ctx)
		case "5":
			h.deleteUser(ctx)
		case "6":
			h.filterUsers(ctx)
		case "7":
			fmt.Fprintln(h.output, "Goodbye!")
			return
		default:
			fmt.Fprintln(h.output, "Invalid menu option. Choose a number from 1 to 7.")
		}
	}
}

func (h *UserHandler) addUser(ctx context.Context) {
	id, ok := h.readID("ID: ")
	if !ok {
		return
	}
	name, ok := h.readLine("Name: ")
	if !ok {
		return
	}
	email, ok := h.readLine("Email: ")
	if !ok {
		return
	}

	err := h.service.Create(ctx, &model.User{ID: id, Name: name, Email: email})
	if err != nil {
		h.renderError(err)
		return
	}

	fmt.Fprintln(h.output, "User added successfully.")
}

func (h *UserHandler) listUsers(ctx context.Context) {
	users, err := h.service.List(ctx)
	if err != nil {
		h.renderError(err)
		return
	}
	h.printUsers(users)
}

func (h *UserHandler) findUser(ctx context.Context) {
	id, ok := h.readID("ID to find: ")
	if !ok {
		return
	}

	user, err := h.service.Find(ctx, id)
	if err != nil {
		h.renderError(err)
		return
	}

	h.printUsers([]*model.User{user})
}

func (h *UserHandler) updateUser(ctx context.Context) {
	id, ok := h.readID("ID to update: ")
	if !ok {
		return
	}
	name, ok := h.readLine("New name: ")
	if !ok {
		return
	}
	email, ok := h.readLine("New email: ")
	if !ok {
		return
	}

	err := h.service.Update(ctx, id, name, email)
	if err != nil {
		h.renderError(err)
		return
	}

	fmt.Fprintln(h.output, "User updated successfully.")
}

func (h *UserHandler) deleteUser(ctx context.Context) {
	id, ok := h.readID("ID to delete: ")
	if !ok {
		return
	}

	err := h.service.Delete(ctx, id)
	if err != nil {
		h.renderError(err)
		return
	}

	fmt.Fprintln(h.output, "User deleted successfully.")
}

func (h *UserHandler) filterUsers(ctx context.Context) {
	term, ok := h.readLine("Search name or email: ")
	if !ok {
		return
	}

	users, err := h.service.Filter(ctx, term)
	if err != nil {
		h.renderError(err)
		return
	}

	h.printUsers(users)
}

// renderError renders error messages in a user-friendly way.
func (h *UserHandler) renderError(err error) {
	if errors.Is(err, domain.ErrUserNotFound) {
		fmt.Fprintln(h.output, "Error: User not found.")
		return
	}

	if errors.Is(err, domain.ErrDuplicateID) {
		fmt.Fprintln(h.output, "Error: A user with that ID already exists.")
		return
	}

	var ve *domain.ValidationError
	if errors.As(err, &ve) {
		fmt.Fprintf(h.output, "Error: Invalid %s: %s\n", ve.Field, ve.Reason)
		return
	}

	// Default error message for unknown errors
	fmt.Fprintln(h.output, "Error: An error occurred.")
}

func (h *UserHandler) readID(prompt string) (int, bool) {
	value, ok := h.readLine(prompt)
	if !ok {
		return 0, false
	}
	id, err := strconv.Atoi(value)
	if err != nil {
		fmt.Fprintln(h.output, "Invalid integer input.")
		return 0, false
	}
	return id, true
}

func (h *UserHandler) readLine(prompt string) (string, bool) {
	fmt.Fprint(h.output, prompt)
	if !h.input.Scan() {
		return "", false
	}
	return strings.TrimSpace(h.input.Text()), true
}

func (h *UserHandler) printUsers(users []*model.User) {
	if len(users) == 0 {
		fmt.Fprintln(h.output, "No users found.")
		return
	}
	fmt.Fprintln(h.output, "ID\tNAME\tEMAIL")
	fmt.Fprintln(h.output, "--\t----\t-----")
	for _, user := range users {
		fmt.Fprintf(h.output, "%d\t%s\t%s\n", user.ID, user.Name, user.Email)
	}
}
