# Database Scanning Feature

## Overview

The DB scanning feature allows you to scan databases present within multiple cloud or on-premise Kubernetes clusters. This feature enables discovery and analysis of database schemas across distributed environments.

## Features

### 1. Multi-Cluster Management (`pkg/cluster`)

- **Register Kubernetes Clusters**: Register both cloud (AWS, GCP, Azure) and on-premise Kubernetes clusters
- **Cluster Connection Management**: Test and maintain connections to multiple clusters
- **Cluster Metadata**: Store cluster information including provider, type, and custom metadata
- **Connection Caching**: Efficient client caching for better performance

### 2. Database Scanner (`pkg/scanner`)

- **Multi-Database Support**: Supports PostgreSQL and MySQL databases
- **Comprehensive Schema Analysis**: Extracts:
  - Tables and views
  - Columns with types, constraints, and defaults
  - Indexes (primary keys, unique indexes, etc.)
  - Foreign key relationships
  - Table statistics (row counts, sizes)
  - Schema information (PostgreSQL)

### 3. API Endpoints

#### Cluster Management

- `POST /api/v1/clusters` - Register a new cluster
- `GET /api/v1/clusters` - List all registered clusters
- `GET /api/v1/clusters/{id}` - Get cluster details
- `DELETE /api/v1/clusters/{id}` - Delete a cluster
- `POST /api/v1/clusters/{id}/test` - Test cluster connection
- `POST /api/v1/clusters/refresh` - Refresh all cluster connections

#### Database Scanning

- `POST /api/v1/scan` - Scan a specific database
- `POST /api/v1/scan/cluster` - Scan all databases in a cluster

## Usage

### Register a Cluster

```bash
curl -X POST http://localhost:8081/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production-cluster",
    "type": "cloud",
    "provider": "aws",
    "kubeconfig": "<base64-encoded-kubeconfig>",
    "metadata": {
      "region": "us-east-1",
      "environment": "production"
    }
  }'
```

### Scan a Database

```bash
curl -X POST http://localhost:8081/api/v1/scan \
  -H "Content-Type: application/json" \
  -d '{
    "cluster_id": "cluster-id",
    "database_host": "db.example.com",
    "database_port": "5432",
    "database_user": "postgres",
    "database_password": "password",
    "database_name": "mydb"
  }'
```

### Scan All Databases in a Cluster

```bash
curl -X POST http://localhost:8081/api/v1/scan/cluster \
  -H "Content-Type: application/json" \
  -d '{
    "cluster_id": "cluster-id",
    "database_password": "default-password"
  }'
```

## Architecture

### Components

1. **Cluster Manager** (`pkg/cluster/manager.go`)
   - Manages multiple Kubernetes cluster connections
   - Handles kubeconfig parsing and client creation
   - Maintains cluster state and connection health

2. **Database Scanner** (`pkg/scanner/scanner.go`)
   - Connects to databases and extracts schema information
   - Supports PostgreSQL and MySQL
   - Returns detailed scan results with table structures

3. **API Handlers**
   - `internal/api/cluster_handler.go` - Cluster management endpoints
   - `internal/api/scanner_handler.go` - Database scanning endpoints

### Integration

The feature is integrated into the manager server (`internal/server/manager.go`):
- Cluster manager and scanner are initialized at server startup
- Routes are registered for cluster and scanning operations
- All endpoints are protected by authentication middleware (if RBAC is enabled)

## Scan Result Structure

```json
{
  "id": "scan-id",
  "cluster_id": "cluster-id",
  "cluster_name": "production-cluster",
  "database_name": "mydb",
  "database_host": "db.example.com",
  "database_port": "5432",
  "database_type": "postgres",
  "status": "success",
  "tables": [
    {
      "name": "users",
      "schema": "public",
      "type": "table",
      "columns": [
        {
          "name": "id",
          "type": "integer",
          "nullable": false,
          "is_primary_key": true
        }
      ],
      "indexes": [...],
      "foreign_keys": [...],
      "row_count": 1000,
      "size_bytes": 8192
    }
  ],
  "table_count": 5,
  "total_row_count": 5000,
  "size_bytes": 40960,
  "scanned_at": "2024-01-01T00:00:00Z",
  "duration_ms": 150
}
```

## Security Considerations

- **Kubeconfig Storage**: Kubeconfigs are stored in memory (not persisted to disk)
- **Password Handling**: Database passwords are passed in requests but not stored
- **RBAC**: All endpoints respect the RBAC configuration
- **Connection Security**: Database connections use SSL when available

## Future Enhancements

- [ ] Persist scan results to a database
- [ ] Schedule periodic scans
- [ ] Compare scan results over time
- [ ] Export scan results to various formats
- [ ] UI components for cluster management and scan visualization
- [ ] Support for additional database types (MongoDB, etc.)
- [ ] Scan result caching
- [ ] Batch scanning with progress tracking

## Dependencies

- `github.com/go-sql-driver/mysql` - MySQL driver
- `github.com/lib/pq` - PostgreSQL driver (already included)
- Kubernetes client-go libraries (already included)

