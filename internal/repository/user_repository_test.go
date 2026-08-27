package repository

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"

	"d10-go-cli-application/internal/domain"
	"d10-go-cli-application/internal/model"
)

func TestMemoryUserRepositoryCreate(t *testing.T) {
	tests := []struct {
		name         string
		user         *model.User
		wantErr      bool
		checkErrType bool
	}{
		{
			name:         "valid user",
			user:         &model.User{ID: 1, Name: "Asha", Email: "asha@example.com"},
			wantErr:      false,
			checkErrType: false,
		},
		{
			name:         "nil user",
			user:         nil,
			wantErr:      true,
			checkErrType: true,
		},
		{
			name:         "empty name",
			user:         &model.User{ID: 1, Name: "", Email: "asha@example.com"},
			wantErr:      true,
			checkErrType: true,
		},
		{
			name:         "empty email",
			user:         &model.User{ID: 1, Name: "Asha", Email: ""},
			wantErr:      true,
			checkErrType: true,
		},
		{
			name:         "whitespace name",
			user:         &model.User{ID: 1, Name: "   ", Email: "asha@example.com"},
			wantErr:      true,
			checkErrType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMemoryUserRepository()
			ctx := context.Background()
			err := repo.Create(ctx, tt.user)

			if (err != nil) != tt.wantErr {
				t.Fatalf("got error %v, want error %v", err != nil, tt.wantErr)
			}

			if tt.wantErr && tt.checkErrType {
				var ve *domain.ValidationError
				if !errors.As(err, &ve) {
					t.Fatalf("expected ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestMemoryUserRepositoryDuplicateID(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	user1 := &model.User{ID: 1, Name: "Asha", Email: "asha@example.com"}
	if err := repo.Create(ctx, user1); err != nil {
		t.Fatal(err)
	}

	user2 := &model.User{ID: 1, Name: "Babu", Email: "babu@example.com"}
	err := repo.Create(ctx, user2)
	if !errors.Is(err, domain.ErrDuplicateID) {
		t.Fatalf("expected ErrDuplicateID, got %v", err)
	}
}

func TestMemoryUserRepositoryList(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	users := []*model.User{
		{ID: 1, Name: "Asha", Email: "asha@example.com"},
		{ID: 2, Name: "Babu", Email: "babu@example.com"},
		{ID: 3, Name: "Charlie", Email: "charlie@example.com"},
	}

	for _, user := range users {
		if err := repo.Create(ctx, user); err != nil {
			t.Fatal(err)
		}
	}

	result, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != len(users) {
		t.Fatalf("expected %d users, got %d", len(users), len(result))
	}

	for i, user := range result {
		if user.ID != users[i].ID {
			t.Fatalf("user %d: expected ID %d, got %d", i, users[i].ID, user.ID)
		}
	}
}

func TestMemoryUserRepositoryFind(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	user := &model.User{ID: 1, Name: "Asha", Email: "asha@example.com"}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatal(err)
	}

	found, err := repo.Find(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	if found.Name != "Asha" {
		t.Fatalf("expected name Asha, got %s", found.Name)
	}
}

func TestMemoryUserRepositoryFindMissing(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	_, err := repo.Find(ctx, 99)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestMemoryUserRepositoryUpdate(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	user := &model.User{ID: 1, Name: "Asha", Email: "asha@example.com"}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatal(err)
	}

	err := repo.Update(ctx, 1, "Asha Rahman", "asha.rahman@example.com")
	if err != nil {
		t.Fatal(err)
	}

	found, err := repo.Find(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	if found.Name != "Asha Rahman" || found.Email != "asha.rahman@example.com" {
		t.Fatalf("expected updated user, got %#v", found)
	}
}

func TestMemoryUserRepositoryUpdateMissing(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	err := repo.Update(ctx, 99, "Nobody", "none@example.com")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestMemoryUserRepositoryUpdateValidation(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	user := &model.User{ID: 1, Name: "Asha", Email: "asha@example.com"}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatal(err)
	}

	err := repo.Update(ctx, 1, "", "new@example.com")
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Field != "name" {
		t.Fatalf("expected field name, got %s", ve.Field)
	}
}

func TestMemoryUserRepositoryDelete(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	users := []*model.User{
		{ID: 1, Name: "A", Email: "a@example.com"},
		{ID: 2, Name: "B", Email: "b@example.com"},
		{ID: 3, Name: "C", Email: "c@example.com"},
	}

	for _, user := range users {
		if err := repo.Create(ctx, user); err != nil {
			t.Fatal(err)
		}
	}

	if err := repo.Delete(ctx, 2); err != nil {
		t.Fatal(err)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 users after deletion, got %d", len(list))
	}

	ids := []int{list[0].ID, list[1].ID}
	if !reflect.DeepEqual(ids, []int{1, 3}) {
		t.Fatalf("expected order [1 3], got %v", ids)
	}
}

func TestMemoryUserRepositoryDeleteMissing(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, 99)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestMemoryUserRepositoryFilter(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	users := []*model.User{
		{ID: 1, Name: "Alice", Email: "alice@example.com"},
		{ID: 2, Name: "Bob", Email: "bob@example.com"},
		{ID: 3, Name: "Charlie", Email: "charlie@alice.com"},
	}

	for _, user := range users {
		if err := repo.Create(ctx, user); err != nil {
			t.Fatal(err)
		}
	}

	// Filter by name
	result, err := repo.Filter(ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 matches for 'alice', got %d", len(result))
	}

	// Filter by email
	result, err = repo.Filter(ctx, "bob@example.com")
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 1 || result[0].ID != 2 {
		t.Fatalf("expected 1 match for 'bob@example.com', got %#v", result)
	}
}

func TestMemoryUserRepositoryDefensiveCopy(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	user := &model.User{ID: 1, Name: "Asha", Email: "asha@example.com"}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatal(err)
	}

	found, err := repo.Find(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	// Modify the returned user
	found.Name = "Modified"

	// Verify the repository's user is unchanged
	original, err := repo.Find(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	if original.Name != "Asha" {
		t.Fatalf("expected original name 'Asha', got '%s'", original.Name)
	}
}

func TestMemoryUserRepositoryConcurrentReads(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	for i := 1; i <= 10; i++ {
		user := &model.User{ID: i, Name: "User" + string(rune(i)), Email: "user" + string(rune(i)) + "@example.com"}
		if err := repo.Create(ctx, user); err != nil {
			t.Fatal(err)
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = repo.List(ctx)
			_, _ = repo.Find(ctx, 1)
			_, _ = repo.Filter(ctx, "User")
		}()
	}

	wg.Wait()
}

func TestMemoryUserRepositoryConcurrentWrites(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	// Pre-populate some users
	for i := 1; i <= 5; i++ {
		user := &model.User{ID: i, Name: "User" + string(rune(i)), Email: "user" + string(rune(i)) + "@example.com"}
		if err := repo.Create(ctx, user); err != nil {
			t.Fatal(err)
		}
	}

	var wg sync.WaitGroup

	// Concurrent updates
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = repo.Update(ctx, id, "Updated"+string(rune(id)), "updated"+string(rune(id))+"@example.com")
		}(i)
	}

	wg.Wait()

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}

	for _, user := range list {
		if !strings.Contains(user.Name, "Updated") {
			t.Fatalf("expected updated name for user %d, got %s", user.ID, user.Name)
		}
	}
}

func TestMemoryUserRepositoryWhitespaceTrimming(t *testing.T) {
	repo := NewMemoryUserRepository()
	ctx := context.Background()

	user := &model.User{ID: 1, Name: "  Asha  ", Email: "  asha@example.com  "}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatal(err)
	}

	found, err := repo.Find(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	if found.Name != "Asha" || found.Email != "asha@example.com" {
		t.Fatalf("expected trimmed user, got Name:%q Email:%q", found.Name, found.Email)
	}
}
