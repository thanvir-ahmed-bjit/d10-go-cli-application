package handler

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"d5/internal/config"
	"d5/internal/model"
	"d5/internal/repository"
)

type UserHandler struct {
	repository *repository.UserRepository
	input      *bufio.Scanner
	output     io.Writer
}

func NewUserHandler(repository *repository.UserRepository, input io.Reader, output io.Writer) *UserHandler {
	return &UserHandler{repository: repository, input: bufio.NewScanner(input), output: output}
}

func (h *UserHandler) Run() {
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
			h.addUser()
		case "2":
			h.listUsers()
		case "3":
			h.findUser()
		case "4":
			h.updateUser()
		case "5":
			h.deleteUser()
		case "6":
			h.filterUsers()
		case "7":
			fmt.Fprintln(h.output, "Goodbye!")
			return
		default:
			fmt.Fprintln(h.output, "Invalid menu option. Choose a number from 1 to 7.")
		}
	}
}

func (h *UserHandler) addUser() {
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
	if err := h.repository.Create(&model.User{ID: id, Name: name, Email: email}); err != nil {
		fmt.Fprintln(h.output, "Error:", err)
		return
	}
	fmt.Fprintln(h.output, "User added successfully.")
}

func (h *UserHandler) listUsers() {
	h.printUsers(h.repository.List())
}

func (h *UserHandler) findUser() {
	id, ok := h.readID("ID to find: ")
	if !ok {
		return
	}
	user, exists := h.repository.Find(id)
	if !exists {
		fmt.Fprintln(h.output, "Error:", repository.ErrUserNotFound)
		return
	}
	h.printUsers([]*model.User{user})
}

func (h *UserHandler) updateUser() {
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
	if err := h.repository.Update(id, name, email); err != nil {
		fmt.Fprintln(h.output, "Error:", err)
		return
	}
	fmt.Fprintln(h.output, "User updated successfully.")
}

func (h *UserHandler) deleteUser() {
	id, ok := h.readID("ID to delete: ")
	if !ok {
		return
	}
	if err := h.repository.Delete(id); err != nil {
		fmt.Fprintln(h.output, "Error:", err)
		return
	}
	fmt.Fprintln(h.output, "User deleted successfully.")
}

func (h *UserHandler) filterUsers() {
	term, ok := h.readLine("Search name or email: ")
	if !ok {
		return
	}
	h.printUsers(h.repository.Filter(term))
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
