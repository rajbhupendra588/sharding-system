# Scripts Directory

This directory contains utility scripts for building, testing, and maintaining the project.

## Available Scripts

### `build.sh`
Builds the backend binaries (router and manager).

**Usage:**
```bash
./scripts/build.sh
```

**Output:**
- `bin/router` - Router service binary
- `bin/manager` - Manager service binary

### `test.sh`
Runs all tests with coverage reporting.

**Usage:**
```bash
./scripts/test.sh
```

**Output:**
- `coverage.out` - Coverage data file
- `coverage.html` - HTML coverage report

### `lint.sh`
Runs code linters using golangci-lint.

**Usage:**
```bash
./scripts/lint.sh
```

**Requirements:**
- golangci-lint (will be installed automatically if not present)

**Note:** Uses `.golangci.yml` configuration file in the project root.

### `clean.sh`
Cleans build artifacts and temporary files.

**Usage:**
```bash
./scripts/clean.sh
```

**Removes:**
- `bin/` directory
- `coverage.out` and `coverage.html`
- `*.tmp` files
- `*.log` files

## Running Scripts

All scripts are executable and can be run directly:

```bash
./scripts/build.sh
```

Or via Makefile (recommended):

```bash
make build-backend    # Uses build.sh
make test-coverage    # Uses test.sh
make lint             # Uses lint.sh
make clean            # Uses clean.sh
```

## Script Standards

All scripts follow these conventions:
- Use `#!/bin/bash` shebang
- Include `set -e` for error handling
- Provide clear output messages
- Exit with appropriate status codes

## Adding New Scripts

When adding new scripts:
1. Make them executable: `chmod +x scripts/new-script.sh`
2. Add shebang: `#!/bin/bash`
3. Add error handling: `set -e`
4. Document in this README
5. Add to Makefile if appropriate

