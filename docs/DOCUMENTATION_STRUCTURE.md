# Documentation Structure

This document describes the organization of all documentation in the `docs/` directory.

## Directory Structure

```
docs/
├── README.md                    # Documentation index (start here)
├── getting-started/             # Getting started guides
│   ├── README.md
│   ├── QUICKSTART.md
│   ├── START_GUIDE.md
│   └── BUILD_AND_START.md
├── api/                         # API documentation
│   ├── README.md
│   ├── API.md
│   └── API_ENDPOINTS.md
├── architecture/                # Architecture and design
│   ├── README.md
│   ├── ARCHITECTURE.md
│   └── SYSTEM_DESIGN.md
├── development/                 # Development guides
│   ├── README.md
│   ├── DEVELOPMENT.md
│   ├── CONTRIBUTING.md
│   ├── IMPLEMENTATION_SUMMARY.md
│   └── scripts.md
├── clients/                     # Client libraries and integration
│   ├── README.md
│   ├── JAVA_QUICKSTART.md
│   ├── QUARKUS_INTEGRATION.md
│   ├── java.md
│   └── quarkus-example.md
└── ui/                          # UI documentation
    ├── README.md
    ├── ARCHITECTURE.md
    ├── QUICKSTART.md
    ├── PERFORMANCE.md
    ├── RESILIENCE.md
    └── IMPROVEMENTS.md
```

## Organization Principles

1. **Single Location**: All documentation is in the `docs/` directory
2. **Categorized**: Documentation is organized by topic/category
3. **Indexed**: Each category has a README.md for navigation
4. **Main Index**: `docs/README.md` provides the main entry point
5. **Consistent**: All documentation follows the same structure

## File Locations

### Root Level
- `README.md` - Main project README (points to docs/)

### Component READMEs
Component-specific READMEs remain in their directories for quick reference:
- `clients/java/README.md` → Moved to `docs/clients/java.md`
- `examples/quarkus-service/README.md` → Moved to `docs/clients/quarkus-example.md`
- `ui/README.md` → Moved to `docs/ui/README.md`
- `scripts/README.md` → Moved to `docs/development/scripts.md`

## Migration Notes

All documentation files have been moved from:
- Root directory → `docs/` subdirectories
- Component directories → `docs/` with appropriate categorization
- Scattered locations → Centralized in `docs/`

## Adding New Documentation

When adding new documentation:

1. **Choose the right category**:
   - Getting started guides → `docs/getting-started/`
   - API docs → `docs/api/`
   - Architecture → `docs/architecture/`
   - Development → `docs/development/`
   - Client libraries → `docs/clients/`
   - UI docs → `docs/ui/`

2. **Update the category README**: Add a link to the new doc in the category's README.md

3. **Update the main index**: Add a link in `docs/README.md` if it's a major document

4. **Follow naming conventions**: Use UPPERCASE.md for major docs, lowercase.md for references

## Benefits

- ✅ All docs in one place
- ✅ Easy to navigate
- ✅ Clear categorization
- ✅ Consistent structure
- ✅ Easy to maintain

