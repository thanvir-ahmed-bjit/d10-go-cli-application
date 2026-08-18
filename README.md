# D5 User Manager

D5 is a standalone interactive Go command-line application for managing users in memory. It demonstrates terminal input, slices, maps, pointers, validation, and formatted output.

## Features

- Add, list, find, update, and delete users.
- Filter users by name or email without regard to case.
- Preserve insertion order while using a map for fast ID lookup.
- Handle invalid input and normal validation errors without crashing.

## Architecture

- `cmd/app`: initializes the repository and starts the CLI.
- `internal/model`: defines the `User` type.
- `internal/repository`: owns the ordered slice and ID map.
- `internal/handler`: reads input, invokes repository methods, and formats output.
- `internal/config`: stores small application constants.

## Requirements and Commands

Go 1.26 or newer is recommended.

```text
cd d5
go run ./cmd/app
go test ./...
```

No database, external package, or existing Go project is required. State exists only for the lifetime of the process.

## Example CLI Session

```text
=== D5 User Manager ===
1. Add user
2. List users
3. Find user by ID
4. Update user
5. Delete user
6. Filter users
7. Exit
Choose an option: 1
ID: 1
Name: Asha
Email: asha@example.com
User added successfully.

Choose an option: 2
ID	NAME	EMAIL
--	----	-----
1	Asha	asha@example.com

Choose an option: 7
Goodbye!
```
