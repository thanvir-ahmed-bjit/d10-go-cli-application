# D10 Concurrent User Manager

D10 is an interactive Go command-line application for managing users with concurrent, object-oriented architecture. It demonstrates idiomatic Go patterns including structs, methods, interfaces, dependency injection, concurrency safety, custom error handling, background logging, and graceful shutdown.

## Features

- Add, list, find, update, and delete users.
- Filter users by name or email (case-insensitive).
- Preserve insertion order while using a map for fast ID lookup.
- Concurrency-safe in-memory repository with sync.RWMutex.
- Background async logger that drains messages during shutdown.
- Context-aware operations supporting cancellation.
- Small consumer-defined interfaces promoting loose coupling.
- Custom domain errors with structured validation errors.
- Comprehensive test coverage with concurrency tests.

## Architecture

The application follows clean architecture principles with clear separation of concerns:

### Directory Structure

- **cmd/app**: Application entry point and composition root
- **internal/config**: Application constants
- **internal/domain**: Domain errors and types
- **internal/handler**: CLI input/output and user interaction layer
- **internal/logger**: Async logger implementation with background goroutine
- **internal/model**: User domain model
- **internal/repository**: User persistence interface and in-memory implementation
- **internal/service**: Business logic layer

### Concurrency Design

The `MemoryUserRepository` uses `sync.RWMutex` to provide thread-safe access:
- Read operations (`List`, `Find`, `Filter`) use `RLock` for concurrent reads
- Write operations (`Create`, `Update`, `Delete`) use `Lock` for exclusive access
- Locks are held for minimal duration with immediate `defer` unlocking

The `AsyncLogger` accepts messages through a buffered channel and processes them in a background goroutine, ensuring:
- Non-blocking message submission (with context cancellation support)
- Proper shutdown that drains accepted messages before exit
- No goroutine leaks or send-to-closed-channel panics

### Custom Error Strategy

- **Sentinel errors** (`ErrDuplicateID`, `ErrUserNotFound`, `ErrLoggerClosed`) for predictable error handling
- **Validation errors** for field-specific validation failures
- **Error wrapping** with `%w` to preserve error chains for inspection with `errors.Is` and `errors.As`

### Logger Lifecycle

1. Logger is created with a buffered channel and buffer size validation
2. Background goroutine starts immediately and ranges over the message channel
3. Messages are logged with timestamps
4. `Close()` atomically closes the message channel and waits for the goroutine to drain remaining messages
5. Safe to call `Close()` multiple times (uses `sync.Once`)
6. Post-shutdown `Log()` calls return `ErrLoggerClosed`

## Requirements

Go 1.26 or newer.

## Running the Application

```bash
cd d10-go-cli-application
go run ./cmd/app
```

## Testing and Validation

### Run all tests

```bash
go test ./...
```

### Run tests with coverage

```bash
go test ./... -cover
```

### Run race detector

```bash
go test -race ./...
```

### Code formatting

```bash
gofmt -w .
```

### Code analysis

```bash
go vet ./...
```

### Complete validation

```bash
gofmt -w .
go test ./...
go vet ./...
go test -race ./...
```

## Example CLI Session

```
=== D10 Concurrent User Manager ===
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

Choose an option: 1
ID: 2
Name: Babu
Email: babu@example.com
User added successfully.

Choose an option: 2
ID	NAME	EMAIL
--	----	-----
1	Asha	asha@example.com
2	Babu	babu@example.com

Choose an option: 3
ID to find: 1
ID	NAME	EMAIL
--	----	-----
1	Asha	asha@example.com

Choose an option: 4
ID to update: 1
New name: Asha Rahman
New email: asha.rahman@example.com
User updated successfully.

Choose an option: 6
Search name or email: rahman
ID	NAME	EMAIL
--	----	-----
1	Asha Rahman	asha.rahman@example.com

Choose an option: 5
ID to delete: 2
User deleted successfully.

Choose an option: 7
Goodbye!

[timestamp] Created user: ID=1, Name=Asha, Email=asha@example.com
[timestamp] Created user: ID=2, Name=Babu, Email=babu@example.com
[timestamp] Updated user: ID=1, Name=Asha Rahman, Email=asha.rahman@example.com
[timestamp] Deleted user: ID=2
```

## Code Quality Standards

- Idiomatic Go using `gofmt` formatting
- Small, focused interfaces defined close to consumers
- Composition over inheritance-style patterns
- No global mutable state
- Defensive copying to prevent external modification of repository state
- No goroutine leaks or data races (verified by race detector)
- Comments explain intent and ownership where necessary

## Design Decisions

### Why sync.RWMutex?
The repository serves both read-heavy (List, Find, Filter) and write operations (Create, Update, Delete). `RWMutex` allows multiple concurrent readers while ensuring exclusive write access.

### Why an AsyncLogger?
The async logger demonstrates proper goroutine lifecycle management, channel-based communication, and graceful shutdown. It prevents blocking the main thread during I/O.

### Why context.Context?
Context is threaded through the handler → service → repository pipeline, enabling:
- Cancellation support for long-running operations
- Timeout propagation
- Proper resource cleanup

### Why small interfaces?
`UserUseCases` (defined by the handler) and `UserRepository` (used by the service) are kept minimal to:
- Reduce coupling between layers
- Make testing easier with focused fakes
- Enable pluggable implementations

## Testing Strategy

### Repository Tests
Validate persistence operations, concurrency safety, defensive copying, and error handling.

### Service Tests
Use fake implementations to test business logic in isolation, verify logging after mutations, and check error wrapping.

### Handler Tests
Use fake service implementations to test CLI interaction, input parsing, error rendering, and menu navigation.

### Logger Tests
Verify message delivery, context cancellation, concurrent usage, proper shutdown, and goroutine cleanup.

--	----	-----
1	Asha	asha@example.com

Choose an option: 7
Goodbye!
```
