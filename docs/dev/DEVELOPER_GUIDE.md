# Developer Guide

## Introduction

This guide is for developers contributing to the Sharding System. It covers setup, code structure, development workflow, testing, and contribution guidelines.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Getting Started](#getting-started)
3. [Project Structure](#project-structure)
4. [Development Workflow](#development-workflow)
5. [Building and Running](#building-and-running)
6. [Testing](#testing)
7. [Code Style](#code-style)
8. [Contributing](#contributing)

## Prerequisites

### Required Tools

- **Go 1.21+**: [Install Go](https://golang.org/doc/install)
- **Docker & Docker Compose**: For local development and testing
- **Node.js 18+**: For UI development (optional)
- **etcd**: Metadata store (can run via Docker)
- **PostgreSQL**: Alternative metadata store or for testing (optional)

### Recommended Tools

- **Git**: Version control
- **Make**: Build automation
- **golangci-lint**: Code linting
- **goimports**: Import formatting

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/your-org/sharding-system.git
cd sharding-system
```

### 2. Install Dependencies

```bash
# Install all dependencies (backend + frontend)
make install-deps

# Or install separately
make install-backend  # Go dependencies
make install-frontend # Node.js dependencies (for UI)
```

### 3. Start Development Environment

```bash
# Start etcd (metadata store)
docker-compose up -d etcd

# Build and start all services
make start-all
```

### 4. Verify Setup

```bash
# Check services are running
curl http://localhost:8080/health  # Router
curl http://localhost:8081/health   # Manager

# Run tests
make test
```

## Project Structure

```
sharding-system/
├── cmd/                    # Main application entry points
│   ├── manager/           # Manager service main
│   └── router/            # Router service main
├── internal/              # Private application code
│   ├── api/               # HTTP handlers
│   │   ├── auth_handler.go
│   │   ├── manager_handler.go
│   │   └── router_handler.go
│   ├── middleware/        # HTTP middleware
│   │   ├── auth.go
│   │   ├── cors.go
│   │   ├── logging.go
│   │   └── recovery.go
│   ├── server/            # Server setup
│   │   ├── manager.go
│   │   └── router.go
│   └── errors/            # Error handling
├── pkg/                   # Public library code
│   ├── catalog/           # Shard catalog management
│   ├── client/            # Go client library
│   ├── config/            # Configuration loading
│   ├── hashing/           # Hash functions
│   ├── manager/           # Manager logic
│   ├── models/            # Data models
│   ├── observability/     # Metrics and logging
│   ├── resharder/         # Resharding operations
│   ├── router/            # Router logic
│   └── security/          # Security (auth, RBAC)
├── ui/                    # Web UI (React + TypeScript)
├── configs/               # Configuration files
│   ├── manager.json
│   └── router.json
├── scripts/               # Build and utility scripts
├── tests/                 # Integration and E2E tests
├── docs/                  # Documentation
└── examples/              # Example code and demos
```

### Key Directories

#### `cmd/`
Contains main entry points for each service:
- `cmd/manager/main.go`: Manager service entry point
- `cmd/router/main.go`: Router service entry point

#### `internal/`
Private application code (not importable by external packages):
- `internal/api/`: HTTP request handlers
- `internal/middleware/`: HTTP middleware (auth, CORS, logging)
- `internal/server/`: Server initialization and setup
- `internal/errors/`: Error handling utilities

#### `pkg/`
Public library code (importable by external packages):
- `pkg/manager/`: Shard management logic
- `pkg/router/`: Query routing logic
- `pkg/catalog/`: Shard catalog operations
- `pkg/hashing/`: Hash computation functions
- `pkg/models/`: Data models and types
- `pkg/client/`: Go client library for external use

## Development Workflow

### Running Individual Services

```bash
# Run Router service
make run-router
# Or directly:
go run ./cmd/router

# Run Manager service
make run-manager
# Or directly:
go run ./cmd/manager

# Run UI development server
make run-frontend
# Or:
cd ui && npm run dev
```

### Building

```bash
# Build all components
make build

# Build backend only
make build-backend

# Build frontend only
make build-frontend
```

Binaries are output to `bin/` directory:
- `bin/router`: Router service binary
- `bin/manager`: Manager service binary

### Code Formatting

```bash
# Format Go code
make fmt

# Or manually:
go fmt ./...
goimports -w .
```

### Linting

```bash
# Run linters
make lint

# Or manually:
golangci-lint run ./...
```

## Building and Running

### Local Development

1. **Start Metadata Store**
   ```bash
   docker-compose up -d etcd
   ```

2. **Set Environment Variables** (if needed)
   ```bash
   export JWT_SECRET="your-secret-key-min-32-chars"
   export USER_DATABASE_DSN="postgresql://user:pass@localhost:5432/users"
   ```

3. **Start Services**
   ```bash
   # Option 1: Use Make
   make start-all

   # Option 2: Run individually
   ./bin/router &
   ./bin/manager &
   cd ui && npm run dev
   ```

### Docker Development

```bash
# Build Docker images
make docker-build

# Start all services via Docker Compose
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

## Testing

### Unit Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests for specific package
go test ./pkg/router/...

# Run tests with verbose output
go test -v ./...
```

### Integration Tests

```bash
# Run integration tests (requires etcd)
go test ./tests/e2e/...
```

### Writing Tests

**Example Unit Test:**
```go
package router_test

import (
    "testing"
    "github.com/sharding-system/pkg/router"
)

func TestGetShardForKey(t *testing.T) {
    r := router.NewRouter(...)
    shardID, err := r.GetShardForKey("user-123")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if shardID == "" {
        t.Error("expected shard ID, got empty string")
    }
}
```

**Best Practices:**
- Use table-driven tests for multiple scenarios
- Test both success and error cases
- Use descriptive test names
- Keep tests independent and isolated
- Mock external dependencies

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Code Style

### Go Code Style

Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines and:

1. **Formatting**: Use `go fmt` and `goimports`
2. **Naming**: Use clear, descriptive names
3. **Comments**: Document exported functions and types
4. **Error Handling**: Always handle errors explicitly
5. **Concurrency**: Use goroutines and channels appropriately

**Example:**
```go
// GetShardForKey returns the shard ID for the given key.
// It computes the hash of the key and looks up the corresponding shard.
func (r *Router) GetShardForKey(key string) (string, error) {
    if key == "" {
        return "", errors.New("key cannot be empty")
    }
    
    hash := r.hasher.Hash(key)
    shardID, err := r.catalog.GetShardForHash(hash)
    if err != nil {
        return "", fmt.Errorf("failed to get shard: %w", err)
    }
    
    return shardID, nil
}
```

### Project-Specific Conventions

1. **Error Handling**: Use the `internal/errors` package for consistent error formatting
2. **Logging**: Use structured logging with `zap.Logger`
3. **Configuration**: Load from JSON files in `configs/` directory
4. **HTTP Handlers**: Keep handlers thin, delegate to service layer

## Contributing

### Contribution Process

1. **Fork the Repository**
   ```bash
   # Fork on GitHub, then:
   git clone https://github.com/your-username/sharding-system.git
   cd sharding-system
   ```

2. **Create a Branch**
   ```bash
   git checkout -b feature/your-feature-name
   # Or:
   git checkout -b fix/your-bug-fix
   ```

3. **Make Changes**
   - Write code following the style guide
   - Add tests for new functionality
   - Update documentation as needed
   - Ensure all tests pass

4. **Commit Changes**
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   ```

   **Commit Message Format:**
   - `feat:` New feature
   - `fix:` Bug fix
   - `docs:` Documentation changes
   - `test:` Test additions/changes
   - `refactor:` Code refactoring
   - `chore:` Maintenance tasks

5. **Push and Create Pull Request**
   ```bash
   git push origin feature/your-feature-name
   ```
   Then create a pull request on GitHub.

### Pull Request Guidelines

- **Description**: Clearly describe what changes were made and why
- **Tests**: Ensure all tests pass
- **Documentation**: Update relevant documentation
- **Breaking Changes**: Clearly mark any breaking changes
- **Review**: Address review comments promptly

### Code Review Checklist

Before submitting a PR, ensure:

- [ ] Code follows style guidelines
- [ ] All tests pass
- [ ] New code has tests
- [ ] Documentation is updated
- [ ] No linter errors
- [ ] Commit messages are clear
- [ ] No sensitive data committed

### Adding New Features

1. **Design**: Discuss major features in an issue first
2. **Implementation**: Follow existing patterns
3. **Testing**: Add comprehensive tests
4. **Documentation**: Update relevant docs
5. **Examples**: Add example code if applicable

### Reporting Bugs

When reporting bugs, include:

- **Description**: Clear description of the issue
- **Steps to Reproduce**: Detailed steps
- **Expected Behavior**: What should happen
- **Actual Behavior**: What actually happens
- **Environment**: OS, Go version, etc.
- **Logs**: Relevant log output

## Debugging

### Debugging Go Code

```bash
# Use Delve debugger
dlv debug ./cmd/router

# Or use VS Code/GoLand debugger
# Set breakpoints and debug
```

### Debugging with Logs

Enable debug logging:
```json
{
  "observability": {
    "log_level": "debug"
  }
}
```

### Common Issues

1. **Port Already in Use**
   - Change port in config file
   - Or stop existing service

2. **etcd Connection Failed**
   - Ensure etcd is running: `docker-compose ps`
   - Check etcd endpoint in config

3. **Authentication Errors**
   - Verify JWT_SECRET is set
   - Check RBAC configuration

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Project Architecture](./architecture/ARCHITECTURE.md)
- [API Reference](../api/API_REFERENCE.md)

## Getting Help

- **Documentation**: Check the [docs](../README.md) directory
- **Issues**: Open an issue on GitHub
- **Discussions**: Use GitHub Discussions for questions

