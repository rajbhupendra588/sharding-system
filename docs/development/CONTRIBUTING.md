# Contributing to Sharding System

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to the project.

## Development Setup

1. **Prerequisites**
   - Go 1.21 or later
   - Docker and Docker Compose
   - Node.js 18+ (for UI development)
   - etcd (or use Docker Compose)

2. **Clone and Setup**
   ```bash
   git clone <repository-url>
   cd sharding-system
   go mod download
   ```

3. **Start Dependencies**
   ```bash
   docker-compose up -d etcd
   ```

## Code Standards

### Go Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` and `goimports` for formatting
- Run `golangci-lint` before submitting PRs
- Write tests for new functionality
- Document exported functions and types

### Project Structure

```
sharding-system/
├── cmd/              # Application entry points
├── internal/         # Private application code
│   ├── api/         # HTTP handlers
│   ├── middleware/  # HTTP middleware
│   ├── server/      # Server setup
│   └── errors/      # Error handling
├── pkg/             # Public packages
├── configs/          # Configuration files
├── scripts/         # Build and utility scripts
└── docs/            # Documentation
```

### Naming Conventions

- Use descriptive names
- Exported types/functions start with capital letter
- Package names should be lowercase, single word
- Constants use PascalCase for exported, camelCase for private

### Error Handling

- Always check errors
- Use `fmt.Errorf` with `%w` for error wrapping
- Return meaningful error messages
- Use the `internal/errors` package for HTTP errors

### Testing

- Write unit tests for all new functionality
- Aim for >80% code coverage
- Use table-driven tests where appropriate
- Test both success and error cases

## Pull Request Process

1. **Create a Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Changes**
   - Write code following the standards above
   - Add tests for new functionality
   - Update documentation as needed

3. **Run Checks**
   ```bash
   # Format code
   go fmt ./...
   goimports -w .

   # Run linter
   ./scripts/lint.sh

   # Run tests
   ./scripts/test.sh
   ```

4. **Commit Changes**
   - Use clear, descriptive commit messages
   - Follow conventional commit format: `type(scope): description`
   - Examples:
     - `feat(router): add connection pooling`
     - `fix(manager): correct shard deletion logic`
     - `docs: update API documentation`

5. **Submit PR**
   - Push your branch
   - Create a pull request with a clear description
   - Reference any related issues
   - Ensure CI checks pass

## Commit Message Format

```
type(scope): subject

body (optional)

footer (optional)
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

## Code Review

- All PRs require at least one approval
- Address review comments promptly
- Keep PRs focused and reasonably sized
- Update documentation for user-facing changes

## Questions?

Feel free to open an issue for questions or discussions about the project.

