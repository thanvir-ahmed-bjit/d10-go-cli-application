package repository

import (
	"errors"
	"reflect"
	"testing"

	"d5/internal/model"
)

func TestUserRepositoryOperations(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*UserRepository)
		test  func(*testing.T, *UserRepository)
	}{
		{
			name: "adding a valid user",
			test: func(t *testing.T, repository *UserRepository) {
				user := &model.User{ID: 1, Name: "Asha", Email: "asha@example.com"}
				if err := repository.Create(user); err != nil {
					t.Fatal(err)
				}
				if got := repository.List(); len(got) != 1 || got[0] != user {
					t.Fatalf("expected user to be stored, got %#v", got)
				}
			},
		},
		{
			name: "rejecting a duplicate ID",
			setup: func(repository *UserRepository) {
				_ = repository.Create(&model.User{ID: 1, Name: "Asha", Email: "asha@example.com"})
			},
			test: func(t *testing.T, repository *UserRepository) {
				err := repository.Create(&model.User{ID: 1, Name: "Babu", Email: "babu@example.com"})
				if !errors.Is(err, ErrDuplicateID) {
					t.Fatalf("expected duplicate ID error, got %v", err)
				}
			},
		},
		{
			name: "finding an existing user",
			setup: func(repository *UserRepository) {
				_ = repository.Create(&model.User{ID: 1, Name: "Asha", Email: "asha@example.com"})
			},
			test: func(t *testing.T, repository *UserRepository) {
				user, ok := repository.Find(1)
				if !ok || user.Name != "Asha" {
					t.Fatalf("expected Asha, got %#v, %v", user, ok)
				}
			},
		},
		{
			name: "finding a missing user",
			test: func(t *testing.T, repository *UserRepository) {
				if user, ok := repository.Find(99); ok || user != nil {
					t.Fatalf("expected missing user, got %#v, %v", user, ok)
				}
			},
		},
		{
			name: "updating an existing user",
			setup: func(repository *UserRepository) {
				_ = repository.Create(&model.User{ID: 1, Name: "Asha", Email: "asha@example.com"})
			},
			test: func(t *testing.T, repository *UserRepository) {
				if err := repository.Update(1, "Asha Rahman", "asha.rahman@example.com"); err != nil {
					t.Fatal(err)
				}
				user, _ := repository.Find(1)
				if user.Name != "Asha Rahman" || user.Email != "asha.rahman@example.com" || repository.List()[0] != user {
					t.Fatalf("slice and map are out of sync: %#v", user)
				}
			},
		},
		{
			name: "updating a missing user",
			test: func(t *testing.T, repository *UserRepository) {
				if err := repository.Update(99, "Nobody", "none@example.com"); !errors.Is(err, ErrUserNotFound) {
					t.Fatalf("expected missing user error, got %v", err)
				}
			},
		},
		{
			name: "deleting a user and preserving order",
			setup: func(repository *UserRepository) {
				_ = repository.Create(&model.User{ID: 1, Name: "A", Email: "a@example.com"})
				_ = repository.Create(&model.User{ID: 2, Name: "B", Email: "b@example.com"})
				_ = repository.Create(&model.User{ID: 3, Name: "C", Email: "c@example.com"})
			},
			test: func(t *testing.T, repository *UserRepository) {
				if err := repository.Delete(2); err != nil {
					t.Fatal(err)
				}
				ids := []int{repository.List()[0].ID, repository.List()[1].ID}
				if !reflect.DeepEqual(ids, []int{1, 3}) {
					t.Fatalf("expected order [1 3], got %v", ids)
				}
				if _, ok := repository.Find(2); ok || len(repository.usersByID) != 2 {
					t.Fatal("map was not synchronized after deletion")
				}
			},
		},
		{
			name: "filtering by name or email",
			setup: func(repository *UserRepository) {
				_ = repository.Create(&model.User{ID: 1, Name: "Asha", Email: "asha@example.com"})
				_ = repository.Create(&model.User{ID: 2, Name: "Babu", Email: "babu@sample.com"})
			},
			test: func(t *testing.T, repository *UserRepository) {
				if got := repository.Filter("ASHA"); len(got) != 1 || got[0].ID != 1 {
					t.Fatalf("name filter returned %#v", got)
				}
				if got := repository.Filter("SAMPLE.COM"); len(got) != 1 || got[0].ID != 2 {
					t.Fatalf("email filter returned %#v", got)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := NewUserRepository()
			if test.setup != nil {
				test.setup(repository)
			}
			test.test(t, repository)
		})
	}
}
