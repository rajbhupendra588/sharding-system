# Standalone DB Sharding Microservice

A production-ready, self-contained database sharding service that provides transparent routing, online resharding, replication management, health monitoring, and comprehensive observability.

## Architecture

```
┌─────────────────────┐
│ Client Microservice │
│  (uses client-lib)  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Shard Router/Proxy │
│    (data plane)     │
└──────────┬──────────┘
           │
    ┌──────┴──────┬──────────┐
    ▼             ▼          ▼
┌─────────┐  ┌─────────┐  ┌─────────┐
│ Shard 1 │  │ Shard 2 │  │ Shard N │
│ (DB)    │  │ (DB)    │  │ (DB)    │
└─────────┘  └─────────┘  └─────────┘
    │
    ▼
┌─────────────────────┐
│  Shard Manager      │
│  (control plane)    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Metadata Store     │
│  (etcd/Postgres)    │
└─────────────────────┘
```

## Components

- **Shard Router/Proxy**: Routes requests to appropriate shards based on shard key
- **Shard Manager**: Control plane API for managing shards, resharding, failover
- **Metadata Store**: Source of truth for shard mappings and routing rules
- **Re-sharder**: Handles online data migration for splits/merges
- **Health Controller**: Monitors shard health and handles failover
- **Client Library**: Lightweight library for microservices to compute shard IDs
- **Web UI**: Production-ready management console for visual administration

## Features

- ✅ Consistent hash-based sharding with virtual nodes
- ✅ Online resharding (split/merge) with minimal downtime
- ✅ Automatic failover and replica promotion
- ✅ Connection pooling and read/write routing
- ✅ Comprehensive metrics and tracing
- ✅ mTLS security and RBAC
- ✅ Audit logging for all operations
- ✅ Health monitoring and alerting

## Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose (for local development)
- etcd (or PostgreSQL for metadata store)
- Node.js 18+ (for UI)

### Build Backend

```bash
go mod download
go build -o bin/shard-router ./cmd/router
go build -o bin/shard-manager ./cmd/manager
go build -o bin/resharder ./cmd/resharder
```

### Run with Docker Compose

```bash
docker-compose up -d
```

### Run Web UI

```bash
cd ui
npm install
npm run dev
```

The UI will be available at `http://localhost:3000`

For detailed setup instructions, see the [Documentation](./docs/README.md).

### Configuration

See `configs/` directory for example configurations.

## Documentation

All documentation is organized in the [`docs/`](./docs/) directory:

- **[Getting Started](./docs/getting-started/QUICKSTART.md)** - Quick start guide
- **[API Documentation](./docs/api/API.md)** - Complete API reference
- **[Architecture](./docs/architecture/ARCHITECTURE.md)** - System architecture
- **[Development Guide](./docs/development/DEVELOPMENT.md)** - Development guidelines
- **[Client Libraries](./docs/clients/JAVA_QUICKSTART.md)** - Client integration guides
- **[UI Documentation](./docs/ui/README.md)** - Web UI documentation

See the [Documentation Index](./docs/README.md) for a complete list of all documentation.

## License

MIT

