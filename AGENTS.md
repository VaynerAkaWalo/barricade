# AGENTS.md - Barricade

Barricade is a Go-based authentication and authorization microservice using DynamoDB for persistence.

## Build Commands

```bash
# Build the application binary
make build

# Build and run the application
make run

# Run all tests
make test                    # Runs: go test ./...

# Run a single test
go test ./test -run TestRegisterHappyPath -v

# Run tests for a specific package
go test ./internal/identity/... -v
go test ./internal/authentication/... -v
go test ./test/... -v
```

## Code Style Guidelines

### Imports
- Group imports into 3 sections separated by blank lines:
  1. Standard library imports
  2. Project-local imports (barricade/*)
  3. Third-party imports
- Use goimports for formatting
- Example:
  ```go
  import (
      "context"
      "encoding/json"
      "net/http"

      "barricade/internal/identity"
      "barricade/pkg/uuid"

      "github.com/VaynerAkaWalo/go-toolkit/xhttp"
      "github.com/aws/aws-sdk-go-v2/aws"
  )
  ```

### Naming Conventions
- **Types**: PascalCase (e.g., `Identity`, `SessionService`)
- **Interfaces**: PascalCase ending with -er (e.g., `Repository`, `SessionService`)
- **Functions/methods**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase (e.g., `secretHash`, `createdAt`)
- **Constants**: PascalCase for exported, camelCase for unexported (e.g., `SessionCookie`)
- **Type aliases for IDs**: Define as string type (e.g., `type Id string`, `type SessionId string`)
- **File names**: snake_case for multi-word files (e.g., `identity_service.go`)
- **Package names**: Short, lowercase, no underscores (e.g., `identity`, `authentication`, `db`)

### Project Structure
```
cmd/           # Application entry points (main.go)
internal/      # Private application code
  authentication/  # Session and auth logic
  db/              # DynamoDB repositories
  identity/        # Identity/user management
  infrastructure/  # HTTP handlers, health checks
pkg/           # Public library code
  uuid/          # UUID utilities
test/          # Integration tests
```

### Type Patterns
- Define domain types with Id as custom string type
- Use struct tags for DynamoDB: `dynamodbav:"fieldName"`
- Use struct tags for JSON: `json:"fieldName"`
- Constructor functions return pointers and errors: `func New(...) (*Type, error)`

### Error Handling
- Use `xhttp.NewError(message, statusCode)` for HTTP-aware errors
- Return errors immediately without wrapping unless adding context
- Log errors using `slog.ErrorContext(ctx, err.Error())`
- Validation errors return 400, not found returns 404, internal errors return 500

### Testing
- Tests are in `test/` package (integration tests) or alongside source (`*_test.go`)
- Use `testcontainers` for DynamoDB integration tests
- Use `stretchr/testify/assert` for assertions
- Test function naming: `Test<Scenario><Condition>` (e.g., `TestRegisterHappyPath`, `TestLoginUnknownUser`)
- Use `t.Cleanup()` for resource cleanup
- Constants for test values: `TEST_NAME`, `TEST_SECRET`

### HTTP Handlers
- Handlers implement `xhttp.RouteHandler` interface
- Define request/response structs for JSON serialization
- Use `xhttp.WriteResponse(w, status, data)` for responses
- Return errors from handlers (handled by xhttp framework)

### Repository Pattern
- Repositories are in `internal/db/` package
- Define interface in domain package, implementation in db package
- Repository methods accept `context.Context` as first parameter
- Use AWS SDK v2 for DynamoDB operations

## Required Environment Variables

```bash
DOMAIN              # Cookie domain
SESSION_TIME        # Session duration in seconds (default: 7200)
DDB_ACCESS_KEY      # AWS access key for DynamoDB
DDB_ACCESS_SECRET_KEY # AWS secret key for DynamoDB
IDENTITY_TABLE_NAME # DynamoDB table name for identities
```

## Dependencies

- Go 1.24+
- AWS SDK v2 (DynamoDB)
- github.com/VaynerAkaWalo/go-toolkit (custom HTTP toolkit)
- testify (testing)
- testcontainers-go (integration tests)
- golang.org/x/crypto (bcrypt for password hashing)
