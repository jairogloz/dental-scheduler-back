# GitHub Copilot Instructions

This file contains instructions for GitHub Copilot to follow when working on this dental scheduler backend project.

## 🚫 Restrictions and Confirmations Required

### File Creation Restrictions

- **Don't create** test files, shell scripts, SQL scripts, Dockerfiles, or markdown files without asking for confirmation first
- **Don't create** migration files without explicit user request and confirmation
- **Don't create** configuration files (docker-compose.yml, Makefile modifications) without permission

### Security Requirements

- **Don't hardcode** sensitive information like passwords, API keys, JWT secrets, or database credentials in committed code
- **Always use** environment variables for configuration values
- **Never commit** actual secrets, tokens, or credentials to version control
- **Use placeholder values** in example files (e.g., `.env.example`)

## 🏗️ Architecture and Project Structure

### Hexagonal/Clean Architecture

This project follows hexagonal (ports and adapters) architecture:

```
internal/
├── domain/           # Core business logic, entities, and interfaces
│   ├── entities/     # Business entities
│   ├── services/     # Domain services
│   └── ports/        # Interfaces (repositories, external services)
├── app/             # Application layer
│   ├── usecases/    # Business use cases and orchestration
│   └── dto/         # Data transfer objects
├── infra/           # Infrastructure layer
│   ├── database/    # Database implementations
│   ├── logger/      # Logging implementation
│   └── config/      # Configuration management
└── http/            # HTTP layer
    ├── handlers/    # HTTP request handlers
    ├── middleware/  # HTTP middleware
    └── routes/      # Route definitions
```

### Layer Dependencies

- **Domain layer**: No dependencies on other layers (pure business logic)
- **Application layer**: Can depend on domain layer only
- **Infrastructure layer**: Can depend on domain and application layers
- **HTTP layer**: Can depend on application layer and infrastructure utilities

## 🔧 Go Best Practices

### Code Style and Structure

- **Use standard Go formatting**: `gofmt` and `goimports` compliant
- **Follow naming conventions**:
  - Exported functions/types: PascalCase
  - Unexported functions/types: camelCase
  - Constants: UPPER_SNAKE_CASE for package-level, camelCase for local
- **Write self-documenting code** with clear function and variable names
- **Add comments** for exported functions, types, and complex logic
- **Keep functions small** and focused on single responsibility

### Error Handling

- **Always handle errors** explicitly, don't ignore them
- **Use custom error types** for domain-specific errors (see `internal/domain/entities/errors.go`)
- **Wrap errors** with context using `fmt.Errorf("operation failed: %w", err)`
- **Return structured error responses** in HTTP handlers with appropriate status codes

### Package Organization

- **Keep packages focused** on single responsibility
- **Use interfaces** to define contracts between layers
- **Avoid circular dependencies** between packages
- **Group related functionality** in the same package

### Database and Repository Pattern

- **Use repository pattern** for data access (see `internal/domain/ports/repositories/`)
- **Implement interfaces** in infrastructure layer
- **Use transactions** for operations that modify multiple entities
- **Handle database errors** appropriately (connection issues, constraint violations)

## 🔐 Authentication and Security

### Supabase Integration

- **Use Supabase JWT middleware** for authentication (`middleware.SupabaseAuth`)
- **Validate tokens** using `SUPABASE_JWT_SECRET` environment variable
- **Extract user information** from JWT claims (user ID, email, role)
- **Apply appropriate middleware** to protected routes

### Route Protection

- **Protected routes**: Use `middleware.SupabaseAuth(logger)` for routes requiring authentication
- **Optional auth routes**: Use `middleware.OptionalAuth(logger)` for routes that benefit from user context
- **Role-based access**: Use `middleware.RequireRole()` for role-specific endpoints

## 📝 HTTP Layer Guidelines

### Handler Implementation

- **Bind request data** using Gin's `ShouldBindJSON()` or `ShouldBindQuery()`
- **Validate input** and return 400 Bad Request for invalid data
- **Log operations** with structured logging using the logger
- **Return consistent response format**:

  ```go
  // Success
  c.JSON(http.StatusOK, gin.H{
      "success": true,
      "data": result,
  })

  // Error
  c.JSON(http.StatusBadRequest, gin.H{
      "success": false,
      "error": gin.H{
          "code": "ERROR_CODE",
          "message": "Human readable message",
      },
  })
  ```

### Error Response Standards

- **Use appropriate HTTP status codes**:
  - 400: Bad Request (validation errors)
  - 401: Unauthorized (authentication required)
  - 403: Forbidden (insufficient permissions)
  - 404: Not Found (resource doesn't exist)
  - 409: Conflict (business rule violations, e.g., appointment conflicts)
  - 500: Internal Server Error (unexpected errors)

### Request/Response DTOs

- **Create specific DTOs** for each endpoint (`internal/app/dto/`)
- **Use validation tags** for input validation
- **Include JSON tags** for proper serialization
- **Convert between DTOs and entities** using conversion functions

## 🗄️ Database and Migration Guidelines

### Migration Best Practices

- **Create reversible migrations** (both up and down)
- **Use descriptive migration names** with timestamps
- **Test migrations** before applying to production
- **Don't modify existing migrations** once they're applied

### SQL Guidelines

- **Use parameterized queries** to prevent SQL injection
- **Handle NULL values** appropriately in database operations
- **Use proper indexes** for query performance
- **Follow PostgreSQL naming conventions** for tables and columns

## 🔍 Testing and Quality

### Code Quality

- **Write unit tests** for business logic in domain and application layers
- **Test error conditions** as well as happy paths
- **Use table-driven tests** for multiple test cases
- **Mock external dependencies** using interfaces

### Logging

- **Use structured logging** with logrus
- **Include relevant context** in log messages (user ID, request ID, etc.)
- **Log at appropriate levels**:
  - Error: Unexpected errors that need attention
  - Warn: Recoverable errors or deprecated usage
  - Info: Important business events
  - Debug: Detailed information for debugging

## 🌍 Environment and Configuration

### Configuration Management

- **Use environment variables** for all configuration
- **Provide default values** for non-sensitive configuration
- **Validate configuration** at application startup
- **Document all environment variables** in `.env.example`

### Required Environment Variables

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=dental_scheduler
DB_USER=username
DB_PASSWORD=password

# Supabase
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your_anon_key
SUPABASE_JWT_SECRET=your_jwt_secret

# Server
SERVER_PORT=8080
SERVER_HOST=localhost

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
```

## 📦 Dependencies and Modules

### Dependency Management

- **Use Go modules** for dependency management
- **Pin dependency versions** for reproducible builds
- **Minimize external dependencies** and prefer standard library when possible
- **Review new dependencies** before adding them

### Common Dependencies Used

- `github.com/gin-gonic/gin`: HTTP web framework
- `github.com/lib/pq`: PostgreSQL driver
- `github.com/sirupsen/logrus`: Structured logging
- `github.com/google/uuid`: UUID generation and handling
- `github.com/golang-jwt/jwt/v5`: JWT token handling
- `github.com/joho/godotenv`: Environment variable loading

## 🚀 Development Workflow

### Code Changes

- **Follow the existing patterns** in the codebase
- **Update interfaces** when adding new methods to repositories or services
- **Update DTOs** when changing API contracts
- **Consider backward compatibility** when modifying existing APIs

### Before Suggesting Code

- **Understand the current architecture** and follow established patterns
- **Check existing implementations** for similar functionality
- **Consider error handling** and edge cases
- **Think about security implications** of changes

Remember: This is a production medical scheduling system. Prioritize security, reliability, and maintainability in all code suggestions.
