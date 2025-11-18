# Testing Guide

This document describes the testing strategy and how to run tests for the sharding system.

## Overview

The project includes comprehensive tests for:
- **Go Backend**: Unit tests, integration tests, and benchmarks
- **Java Client**: Unit tests using JUnit 5
- **TypeScript Frontend**: Component tests and utility tests using Vitest

## Running Tests

### Go Tests

Run all Go tests:
```bash
make test
```

Run tests with coverage:
```bash
make test-coverage
```

Run tests for a specific package:
```bash
go test ./pkg/hashing/... -v
go test ./pkg/router/... -v
go test ./internal/errors/... -v
```

### Java Tests

Run Java client tests:
```bash
cd clients/java
mvn test
```

### Frontend Tests

Run frontend tests:
```bash
cd ui
npm test
```

Run tests in watch mode:
```bash
cd ui
npm test -- --watch
```

## Test Coverage

### Current Coverage

- **pkg/hashing**: Comprehensive unit tests covering all hash functions and consistent hashing
- **pkg/router**: Unit tests with mock catalog
- **pkg/manager**: Unit tests with mock catalog and resharder
- **internal/errors**: Complete test coverage for error handling
- **Frontend**: Tests for HTTP client and utility functions

### Coverage Goals

- Target: 80%+ code coverage for critical packages
- Current focus: Core routing and sharding logic
- Future: Integration tests for end-to-end scenarios

## Writing Tests

### Go Test Guidelines

1. **Naming**: Test files must end with `_test.go`
2. **Functions**: Test functions must start with `Test`
3. **Benchmarks**: Benchmark functions must start with `Benchmark`
4. **Table-driven tests**: Use table-driven tests for multiple scenarios

Example:
```go
func TestHashFunction(t *testing.T) {
    tests := []struct {
        name string
        input string
        expected uint64
    }{
        {"empty string", "", 0},
        {"test key", "test-key", 12345},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Frontend Test Guidelines

1. **Use Vitest**: All frontend tests use Vitest
2. **Mock external dependencies**: Mock HTTP clients, APIs, etc.
3. **Test user interactions**: Use React Testing Library for component tests

Example:
```typescript
import { describe, it, expect, vi } from 'vitest';

describe('Component', () => {
  it('should render correctly', () => {
    // test implementation
  });
});
```

## Continuous Integration

Tests run automatically on:
- Push to `main` or `develop` branches
- Pull requests
- Scheduled nightly builds

See `.github/workflows/ci.yml` for CI configuration.

## Test Data

- Use mocks for external dependencies (databases, APIs)
- Use test fixtures for complex data structures
- Clean up test data after each test

## Performance Testing

Benchmarks are included for critical paths:
```bash
go test -bench=. ./pkg/hashing/...
```

## Troubleshooting

### Go Tests Failing

1. Ensure all dependencies are installed: `go mod download`
2. Check for linting errors: `make lint`
3. Verify Go version: `go version` (requires Go 1.21+)

### Frontend Tests Failing

1. Install dependencies: `cd ui && npm install`
2. Check Node version: `node --version` (requires Node 18+)
3. Clear cache: `rm -rf ui/node_modules ui/.vite`

### Java Tests Failing

1. Ensure Maven is installed: `mvn --version`
2. Check JDK version: `java -version` (requires JDK 11+)
3. Clean and rebuild: `mvn clean test`

