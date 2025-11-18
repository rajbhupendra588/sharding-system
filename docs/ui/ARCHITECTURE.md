# Architecture Documentation

## Overview

This UI follows a **feature-based, modular architecture** emphasizing separation of concerns, testability, maintainability, and scalability. The architecture is designed for production use with focus on performance, resilience, and low latency.

## Directory Structure

```
ui/src/
├── core/                    # Core application infrastructure
│   ├── config/             # Configuration and constants
│   │   ├── constants.ts    # Application constants
│   │   ├── app-config.ts   # Runtime configuration manager
│   │   └── index.ts        # Config exports
│   └── http/               # HTTP client infrastructure
│       ├── client.ts        # Base HTTP client
│       ├── api-factory.ts   # API client factory
│       ├── retry.ts         # Retry logic
│       ├── cache.ts         # Request caching
│       └── index.ts         # HTTP exports
│
├── features/                # Feature modules (domain-driven)
│   ├── shard/              # Shard management feature
│   │   ├── types/          # Feature-specific types
│   │   ├── services/        # Data access layer (repositories)
│   │   ├── hooks/          # Business logic hooks
│   │   └── index.ts        # Feature exports
│   ├── query/              # Query execution feature
│   ├── resharding/         # Resharding operations feature
│   └── health/             # Health monitoring feature
│
├── shared/                  # Shared code across features
│   ├── types/              # Shared type definitions
│   ├── utils/              # Shared utility functions
│   │   ├── formatting.ts   # Formatting utilities
│   │   ├── validation.ts   # Validation utilities
│   │   ├── status.ts       # Status helpers
│   │   └── debounce.ts     # Debounce/throttle
│   ├── lib/                # Shared libraries
│   │   └── cn.ts           # Class name utility
│   └── index.ts            # Shared exports
│
├── components/              # Reusable UI components
│   ├── ui/                 # Base UI components
│   └── Layout.tsx          # Layout components
│
├── pages/                   # Page components
│   ├── Dashboard.tsx
│   ├── Shards.tsx
│   └── ...
│
└── store/                   # Global state management
    └── auth-store.ts        # Authentication store
```

## Architecture Principles

### 1. Feature-Based Organization

Each feature is self-contained with:
- **Types**: Domain-specific type definitions
- **Services**: Data access layer (repositories)
- **Hooks**: Business logic and React Query integration
- **Components**: Feature-specific UI components (if needed)

**Benefits**:
- Clear boundaries between features
- Easy to locate feature-related code
- Facilitates code splitting
- Enables team ownership

### 2. Separation of Concerns

#### Layers:

1. **Presentation Layer** (`pages/`, `components/`)
   - UI components
   - User interactions
   - Visual representation

2. **Business Logic Layer** (`features/*/hooks/`)
   - Custom hooks encapsulate business logic
   - React Query integration
   - State management
   - Side effects

3. **Data Access Layer** (`features/*/services/`)
   - Repository pattern
   - API communication
   - Data transformation
   - Error handling

4. **Infrastructure Layer** (`core/`)
   - HTTP client
   - Configuration
   - Cross-cutting concerns

### 3. Repository Pattern

Each feature has a repository that handles all data access:

```typescript
// features/shard/services/shard-repository.ts
export class ShardRepository {
  async findAll(): Promise<Shard[]> { ... }
  async findById(id: string): Promise<Shard> { ... }
  async create(request: CreateShardRequest): Promise<Shard> { ... }
  async delete(id: string): Promise<void> { ... }
}
```

**Benefits**:
- Single responsibility
- Easy to mock for testing
- Can swap implementations
- Centralized error handling

### 4. Custom Hooks for Business Logic

Business logic is encapsulated in custom hooks:

```typescript
// features/shard/hooks/use-shards.ts
export function useShards() {
  return useQuery({
    queryKey: ['shards'],
    queryFn: () => shardRepository.findAll(),
    ...
  });
}

export function useCreateShard() {
  return useMutation({
    mutationFn: (data) => shardRepository.create(data),
    onSuccess: () => { ... },
    ...
  });
}
```

**Benefits**:
- Reusable business logic
- Consistent patterns
- Easy to test
- Type-safe

### 5. Type Organization

Types are organized by domain:
- **Shared Types** (`shared/types/`): Common types used across features
- **Feature Types** (`features/*/types/`): Domain-specific types
- **Core Types**: Infrastructure types

**Benefits**:
- Clear type ownership
- Prevents circular dependencies
- Easy to find types
- Better IntelliSense

### 6. Configuration Management

Centralized configuration:

```typescript
// core/config/constants.ts
export const API_CONFIG = { ... };
export const REFRESH_INTERVALS = { ... };

// core/config/app-config.ts
export class AppConfigManager {
  getConfig(): AppConfig { ... }
  updateConfig(updates: Partial<AppConfig>): void { ... }
}
```

**Benefits**:
- Single source of truth
- Runtime configuration
- Easy to test
- Environment-specific configs

## Performance Optimizations

### 1. Request Caching
- Aggressive caching for read operations
- Cache invalidation on mutations
- Stale-while-revalidate pattern

### 2. Request Deduplication
- Automatic deduplication of identical requests
- Shared request state across components

### 3. Optimistic Updates
- Immediate UI updates for mutations
- Rollback on error

### 4. Code Splitting
- Route-based code splitting
- Feature-based lazy loading
- Dynamic imports for heavy components

### 5. Memoization
- React.memo for expensive components
- useMemo for computed values
- useCallback for stable function references

## Resilience Features

### 1. Error Retry Logic
- Exponential backoff
- Configurable retry attempts
- Circuit breaker pattern

### 2. Error Boundaries
- Component-level error boundaries
- Graceful degradation
- Error recovery

### 3. Offline Support
- Service worker for offline capability
- Request queuing when offline
- Sync when online

### 4. Health Monitoring
- Automatic health checks
- Degraded mode detection
- Fallback strategies

## Data Flow

```
User Action
    ↓
Page Component
    ↓
Custom Hook (useShards, useCreateShard, etc.)
    ↓
Repository (shardRepository, queryRepository, etc.)
    ↓
HTTP Client (HttpClient)
    ↓
API Endpoint
```

## Example: Adding a New Feature

### 1. Create Feature Structure

```bash
src/features/my-feature/
├── types/
│   └── index.ts
├── services/
│   └── my-feature-repository.ts
├── hooks/
│   └── use-my-feature.ts
└── index.ts
```

### 2. Define Types

```typescript
// features/my-feature/types/index.ts
export interface MyFeature {
  id: string;
  name: string;
}
```

### 3. Create Repository

```typescript
// features/my-feature/services/my-feature-repository.ts
import { ApiFactory } from '@/core/http';
import type { MyFeature } from '../types';

export class MyFeatureRepository {
  private client = ApiFactory.getManagerClient();

  async findAll(): Promise<MyFeature[]> {
    return this.client.get<MyFeature[]>('/my-feature');
  }
}

export const myFeatureRepository = new MyFeatureRepository();
```

### 4. Create Hooks

```typescript
// features/my-feature/hooks/use-my-feature.ts
import { useQuery } from '@tanstack/react-query';
import { myFeatureRepository } from '../services/my-feature-repository';

export function useMyFeature() {
  return useQuery({
    queryKey: ['my-feature'],
    queryFn: () => myFeatureRepository.findAll(),
  });
}
```

### 5. Export Feature

```typescript
// features/my-feature/index.ts
export * from './types';
export * from './services/my-feature-repository';
export * from './hooks/use-my-feature';
```

### 6. Use in Component

```typescript
// pages/MyFeaturePage.tsx
import { useMyFeature } from '@/features/my-feature';

export default function MyFeaturePage() {
  const { data, isLoading } = useMyFeature();
  // ...
}
```

## Best Practices

### 1. Feature Independence
- Features should not depend on other features
- Use shared module for common code
- Import from feature index, not internal files

### 2. Type Safety
- Always define types for API responses
- Use TypeScript strict mode
- Avoid `any` types

### 3. Error Handling
- Handle errors at repository level
- Provide user feedback via hooks
- Use error boundaries for component errors

### 4. Testing
- Test repositories independently
- Test hooks with React Testing Library
- Mock repositories in component tests

### 5. Code Organization
- One file per class/function
- Group related exports in index files
- Use barrel exports for public API

## Migration Guide

### From Old Structure

**Old**:
```typescript
import { apiClient } from '@/lib/api-client';
const shards = await apiClient.listShards();
```

**New**:
```typescript
import { useShards } from '@/features/shard';
const { data: shards } = useShards();
```

### Legacy Support

The old `apiClient` is still available for backward compatibility but is deprecated. It wraps the new repositories internally.

## Benefits of This Architecture

1. **Maintainability**: Clear structure makes code easy to navigate
2. **Scalability**: Easy to add new features without affecting existing ones
3. **Testability**: Each layer can be tested independently
4. **Reusability**: Hooks and repositories can be reused across components
5. **Type Safety**: Strong typing throughout the application
6. **Performance**: Code splitting by feature enables lazy loading
7. **Team Collaboration**: Clear ownership boundaries

## References

- [Feature-Sliced Design](https://feature-sliced.design/)
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Repository Pattern](https://martinfowler.com/eaaCatalog/repository.html)
- [React Query Best Practices](https://tanstack.com/query/latest/docs/react/guides/best-practices)

