# AGENTS.md - Barricade

Barricade is a Go-based authentication and authorization microservice using DynamoDB for persistence.

## Dependencies

- Go 1.24+
- AWS SDK v2 (DynamoDB)
- github.com/VaynerAkaWalo/go-toolkit (custom HTTP toolkit)
- testify (testing)
- testcontainers-go (integration tests)
- golang.org/x/crypto (bcrypt for password hashing)

## Code Style Guidelines

### General
- Do not add comments to code
- If service function requires more than 2 arguments (excluding context) create struct for params with name `XXXparams`

### Naming Conventions
- **File names**: snake_case for multi-word files (e.g., `identity_service.go`)
- **Package names**: Short, lowercase, no underscores
  - **Domain Packages**:  (e.g., `identity`, `authentication`, `db`)
  - **Infra / Utils**: prefixed with i for  (e.g., `ihttp`, `ictx`)
- branch name should start with either `feat`, `chore` or `fix`

### Type Patterns
- Domain objects should be defined in domain package in `DOMAIN_OBJECT_NAME.go`.
- Service type, repository interface and service methods should be defined in service class
- Fields representing fixed set of values OR domain ids should be declared as custom type (e.g. `identityId`, `keyType`, `algorithm`)
- Use struct tags for JSON: `json:"fieldName"`

### Error Handling
- Every domain should declare domain errors in errors.go
- Handler should map domain and internal errors to http errors `xhttp.NewError(message, statusCode)`
- Log errors using `slog.ErrorContext(ctx, err.Error())`

### Testing
- Use `stretchr/testify/assert` for assertions
- Tests are in `test/` package (integration tests) or alongside source (`*_test.go`)
- Use `testcontainers` for DynamoDB integration tests
- Test function naming: `Test<Scenario><Condition>` (e.g., `TestRegisterHappyPath`, `TestLoginUnknownUser`)
- Use `t.Cleanup()` for resource cleanup

### HTTP Handlers
- Handlers implement `xhttp.RouteHandler` interface
- Define request/response structs in Handler for JSON serialization
- Use `xhttp.WriteResponse(w, status, data)` for responses
- Return errors from handlers (handled by xhttp framework)

### Repository Pattern
- Repositories are implemented in domain package
- Define interface in service
- Repository methods accept `context.Context` as first parameter
- Use AWS SDK v2 for DynamoDB operations
