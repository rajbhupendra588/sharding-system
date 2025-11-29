# Documentation Structure Overview

This document provides an overview of the documentation structure and organization.

## Directory Structure

```
docs/
├── README.md                    # Main documentation index
├── STRUCTURE.md                 # This file - structure overview
│
├── admin/                       # Administrator Documentation
│   └── ADMIN_SETUP_AND_MULTI_TENANCY.md
│
├── api/                         # API Documentation
│   └── API_REFERENCE.md
│
├── architecture/                # Architecture & Design Documentation
│   ├── ARCHITECTURE.md
│   ├── MULTI_TENANCY.md
│   ├── SYSTEM_DESIGN.md
│   └── database_sharding_guide.md
│
├── changelog/                   # Version History & Phase Completion
│   ├── PHASE1_COMPLETE.md
│   ├── PHASE2_COMPLETE.md
│   └── PHASE2_IMPLEMENTATION.md
│
├── customer/                    # Customer-Facing Documentation
│   └── COST_AND_LICENSE_INFO.md
│
├── deployment/                  # Deployment & Operations
│   ├── DEPLOYMENT_GUIDE.md
│   ├── K8S_INTEGRATION_GUIDE.md
│   └── K8S_REQUIREMENTS.md
│
├── dev/                         # Developer Documentation
│   ├── DEVELOPER_GUIDE.md
│   ├── SECURITY.md
│   └── TESTING_GUIDE.md
│
├── features/                    # Feature-Specific Documentation
│   └── DB_SCANNING_FEATURE.md
│
├── product/                     # Product & Business Documentation
│   ├── COMPETITIVE_STRATEGY.md
│   ├── IMPLEMENTATION_GUIDE.md
│   ├── MARKET_STRATEGY.md
│   ├── PRODUCT_OVERVIEW.md
│   └── ROADMAP.md
│
├── security/                     # Security & Compliance
│   ├── DEPLOYMENT_MODEL.md
│   └── SECURITY_AND_COMPLIANCE.md
│
├── swagger/                     # Swagger/OpenAPI Documentation
│   ├── manager/
│   │   ├── manager_docs.go
│   │   ├── manager_swagger.json
│   │   └── manager_swagger.yaml
│   └── router/
│       ├── router_docs.go
│       ├── router_swagger.json
│       └── router_swagger.yaml
│
└── user/                        # User Documentation
    ├── CLIENT_APPLICATIONS.md
    ├── CONFIGURATION_GUIDE.md
    ├── FAQ.md
    ├── FINDING_APPLICATIONS.md
    ├── RELEASE_NOTES.md
    ├── SETUP_GUIDE.md
    └── USER_GUIDE.md
```

## Organization Principles

### 1. **User-Focused Organization**
- **user/**: End-user guides, setup, configuration, FAQs
- **admin/**: Administrative setup and multi-tenancy
- **customer/**: Customer-facing information (pricing, licensing)

### 2. **Technical Documentation**
- **dev/**: Developer guides, testing, security practices
- **architecture/**: System design, architecture patterns, technical concepts
- **api/**: API reference documentation
- **swagger/**: Interactive API documentation

### 3. **Operations & Deployment**
- **deployment/**: Deployment guides, Kubernetes integration, requirements
- **features/**: Feature-specific documentation

### 4. **Business & Product**
- **product/**: Product overview, strategy, roadmap
- **changelog/**: Version history and phase completion summaries

### 5. **Security & Compliance**
- **security/**: Security documentation, compliance, deployment models

## Navigation Guide

### For New Users
1. Start with [User Guide](./user/USER_GUIDE.md)
2. Follow [Setup Guide](./user/SETUP_GUIDE.md)
3. Review [FAQ](./user/FAQ.md) for common questions

### For Administrators
1. Read [Admin Setup and Multi-Tenancy](./admin/ADMIN_SETUP_AND_MULTI_TENANCY.md)
2. Review [Deployment Guide](./deployment/DEPLOYMENT_GUIDE.md)
3. Check [Kubernetes Integration](./deployment/K8S_INTEGRATION_GUIDE.md)

### For Developers
1. Start with [Developer Guide](./dev/DEVELOPER_GUIDE.md)
2. Review [Architecture](./architecture/ARCHITECTURE.md)
3. Check [API Reference](./api/API_REFERENCE.md)
4. Read [Testing Guide](./dev/TESTING_GUIDE.md)

### For Operations Teams
1. Review [Deployment Guide](./deployment/DEPLOYMENT_GUIDE.md)
2. Check [Kubernetes Requirements](./deployment/K8S_REQUIREMENTS.md)
3. Read [Security and Compliance](./security/SECURITY_AND_COMPLIANCE.md)

## File Naming Conventions

- **UPPERCASE_WITH_UNDERSCORES.md**: Main documentation files
- **lowercase_with_underscores.md**: Supporting or feature-specific files
- All files use `.md` extension
- Descriptive names that indicate content

## Link Guidelines

When referencing other documentation:
- Use relative paths: `../deployment/DEPLOYMENT_GUIDE.md`
- Always use forward slashes `/`
- Include descriptive link text: `[Deployment Guide](../deployment/DEPLOYMENT_GUIDE.md)`
- Update links when moving files

## Maintenance

- Keep the [README.md](./README.md) updated with new documents
- Maintain consistent structure across directories
- Update cross-references when reorganizing
- Review structure periodically for improvements




