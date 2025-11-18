# Startup Guide

## Prerequisites Check

Before starting, ensure you have:

1. **Go 1.21+** installed and in PATH
   ```bash
   go version
   ```

2. **Node.js 18+** installed
   ```bash
   node --version
   npm --version
   ```

3. **Docker Desktop** running (for etcd)
   ```bash
   docker ps
   ```

## Quick Start Options

### Option 1: Start Everything Automatically (Recommended)

```bash
# This will start backend and frontend
./scripts/start-all.sh
```

### Option 2: Start Services Separately

#### Start Backend Only
```bash
./scripts/start-backend.sh
```

This will:
- Start etcd via Docker
- Build router and manager services
- Start router on port 8080
- Start manager on port 8081

#### Start Frontend Only
```bash
./scripts/start-frontend.sh
```

This will:
- Install npm dependencies (if needed)
- Start Vite dev server on port 3000

### Option 3: Manual Start

#### 1. Start etcd
```bash
docker-compose up -d etcd
```

#### 2. Install Backend Dependencies
```bash
go mod download
go mod tidy
```

#### 3. Build Backend
```bash
make build-backend
# or
go build -o bin/router ./cmd/router
go build -o bin/manager ./cmd/manager
```

#### 4. Start Router (Terminal 1)
```bash
./bin/router
# or
go run ./cmd/router
```

#### 5. Start Manager (Terminal 2)
```bash
./bin/manager
# or
go run ./cmd/manager
```

#### 6. Install Frontend Dependencies
```bash
cd ui
npm install
```

#### 7. Start Frontend (Terminal 3)
```bash
cd ui
npm run dev
```

## Access the Application

Once all services are running:

- **Frontend UI**: http://localhost:3000
- **Router API**: http://localhost:8080
- **Manager API**: http://localhost:8081
- **etcd**: http://localhost:2379

## Verify Services

### Check Backend Services
```bash
# Check router
curl http://localhost:8080/health

# Check manager
curl http://localhost:8081/health

# List shards
curl http://localhost:8081/api/v1/shards
```

### Check Frontend
Open browser: http://localhost:3000

## Stop Services

### Stop All Services
```bash
./scripts/stop-all.sh
```

### Stop Backend Only
```bash
./scripts/stop-backend.sh
```

### Stop Frontend
Press `Ctrl+C` in the terminal running the frontend

### Stop etcd
```bash
docker-compose stop etcd
```

## Troubleshooting

### Go Not Found
```bash
# Add Go to PATH (example for macOS)
export PATH=$PATH:/usr/local/go/bin

# Or install Go from https://go.dev
```

### Docker Not Running
```bash
# Start Docker Desktop application
# Or start Docker daemon:
sudo systemctl start docker  # Linux
```

### Port Already in Use
```bash
# Check what's using the port
lsof -i :8080  # Router
lsof -i :8081  # Manager
lsof -i :3000  # Frontend

# Kill the process or change ports in config files
```

### Frontend Dependencies Issues
```bash
cd ui
rm -rf node_modules package-lock.json
npm install
```

### Backend Build Issues
```bash
# Clean and rebuild
rm -rf bin/
go mod tidy
go build -o bin/router ./cmd/router
go build -o bin/manager ./cmd/manager
```

## Using Make Commands

```bash
# Install all dependencies
make install-deps

# Build everything
make build

# Build backend only
make build-backend

# Build frontend only
make build-frontend

# Run router
make run-router

# Run manager
make run-manager

# Run frontend
make run-frontend
```

## Docker Compose (Alternative)

You can also use Docker Compose for everything:

```bash
# Start all services (including backend in containers)
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all
docker-compose down
```

