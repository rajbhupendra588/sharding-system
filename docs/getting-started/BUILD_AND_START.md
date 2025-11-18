# Build and Start Guide

## ✅ Current Status

- ✅ Frontend dependencies installed
- ⚠️  Backend: Go needs to be in PATH
- ⚠️  Docker: Needs to be running for etcd

## Quick Start Commands

### 1. Start Backend Services

**Option A: Using the startup script**
```bash
./scripts/start-backend.sh
```

**Option B: Manual steps**
```bash
# Step 1: Start etcd (requires Docker running)
docker-compose up -d etcd

# Step 2: Install Go dependencies
go mod download
go mod tidy

# Step 3: Build backend services
go build -o bin/router ./cmd/router
go build -o bin/manager ./cmd/manager

# Step 4: Start router (in one terminal)
./bin/router

# Step 5: Start manager (in another terminal)
./bin/manager
```

### 2. Start Frontend UI

**Option A: Using the startup script**
```bash
./scripts/start-frontend.sh
```

**Option B: Manual**
```bash
cd ui
npm run dev
```

The frontend will be available at: **http://localhost:3000**

### 3. Start Everything at Once

```bash
./scripts/start-all.sh
```

## Service URLs

Once started, access:

- **Frontend UI**: http://localhost:3000
- **Router API**: http://localhost:8080
- **Manager API**: http://localhost:8081
- **Metrics (Router)**: http://localhost:9090/metrics
- **Metrics (Manager)**: http://localhost:9091/metrics

## Verify Services Are Running

```bash
# Check router health
curl http://localhost:8080/health

# Check manager health
curl http://localhost:8081/health

# List shards
curl http://localhost:8081/api/v1/shards
```

## Prerequisites Setup

### If Go is not in PATH:

**macOS:**
```bash
# Add to ~/.zshrc or ~/.bash_profile
export PATH=$PATH:/usr/local/go/bin

# Then reload
source ~/.zshrc  # or source ~/.bash_profile
```

**Linux:**
```bash
# Add to ~/.bashrc
export PATH=$PATH:/usr/local/go/bin

# Then reload
source ~/.bashrc
```

### If Docker is not running:

**macOS:**
- Open Docker Desktop application

**Linux:**
```bash
sudo systemctl start docker
```

## Using Make Commands

```bash
# Install all dependencies
make install-deps

# Build backend
make build-backend

# Build frontend
make build-frontend

# Build everything
make build

# Run router
make run-router

# Run manager
make run-manager

# Run frontend
make run-frontend
```

## Stop Services

```bash
# Stop everything
./scripts/stop-all.sh

# Stop backend only
./scripts/stop-backend.sh

# Stop frontend: Press Ctrl+C in the terminal
```

## Troubleshooting

### Port Already in Use
```bash
# Find and kill process on port 8080
lsof -ti:8080 | xargs kill -9

# Or change port in configs/router.json
```

### etcd Connection Error
```bash
# Make sure etcd is running
docker ps | grep etcd

# Restart etcd
docker-compose restart etcd
```

### Frontend Build Errors
```bash
cd ui
rm -rf node_modules
npm install
npm run dev
```

## Next Steps

1. **Start Docker Desktop** (if not running)
2. **Ensure Go is in PATH** (check with `go version`)
3. **Run `./scripts/start-all.sh`** to start everything
4. **Open http://localhost:3000** in your browser

For detailed documentation, see:
- [START_GUIDE.md](START_GUIDE.md) - Detailed startup guide
- [QUICKSTART.md](QUICKSTART.md) - Quick start with examples
- [docs/API.md](docs/API.md) - API documentation

