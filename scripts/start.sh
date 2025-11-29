#!/usr/bin/env bash
# Start All Sharding System Services
# This script starts etcd, manager, router and frontend services
# Usage: ./scripts/start.sh [--force-restart]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# Parse args
FORCE_RESTART=false
for arg in "$@"; do
    case "$arg" in
        --force-restart) FORCE_RESTART=true ;;
        -h|--help)
            cat <<EOF
Usage: $0 [--force-restart]
  --force-restart   Forcefully stop, rebuild and restart all services (docker compose down + rebuild + restart)
EOF
            exit 0
            ;;
        *) echo "Unknown arg: $arg"; exit 2 ;;
    esac
done

# Ensure helper dirs exist
mkdir -p logs bin

# Choose docker compose command
DOCKER_COMPOSE_CMD="docker-compose"
if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE_CMD="docker compose"
fi

echo "🚀 Starting Sharding System Services... (force-restart=${FORCE_RESTART})"
echo ""

# Utility: safe kill pids list
safe_kill_pids() {
    local pids="$1"
    if [ -n "$pids" ]; then
        echo "   Killing PIDs: $pids"
        # shellcheck disable=SC2086
        # Use xargs without -r (not available on macOS) and handle empty input
        for pid in $pids; do
            kill -9 "$pid" 2>/dev/null || true
        done
        sleep 1
    fi
}

# If force restart requested, attempt to stop docker compose services (etcd) and remove containers
if [ "$FORCE_RESTART" = true ]; then
    echo "⚡ --force-restart given: attempting to stop docker compose services and remove containers..."
    if command -v docker >/dev/null 2>&1; then
        # best-effort: stop compose services defined for etcd (and the rest if you want)
        if $DOCKER_COMPOSE_CMD ps >/dev/null 2>&1; then
            echo "   Running: $DOCKER_COMPOSE_CMD down --remove-orphans"
            $DOCKER_COMPOSE_CMD down --remove-orphans || {
                echo "   ⚠️  docker compose down failed (continuing)"
            }
        else
            echo "   No compose project detected or docker not ready - skipping compose down"
        fi

        # remove any existing sharding-etcd container forcefully if present
        if docker ps -a --format '{{.Names}}' | grep -q "^sharding-etcd$"; then
            echo "   Removing existing sharding-etcd container..."
            docker rm -f sharding-etcd 2>/dev/null || true
        fi
    else
        echo "   docker not available - cannot stop compose services"
    fi

    # Remove stale pid files and kill processes for manager/router/frontend
    echo "   Removing stale pid files and killing processes..."
    [ -f .manager.pid ] && safe_kill_pids "$(cat .manager.pid 2>/dev/null || true)" && rm -f .manager.pid
    [ -f .router.pid ]  && safe_kill_pids "$(cat .router.pid 2>/dev/null || true)" && rm -f .router.pid
    [ -f .frontend.pid ]&& safe_kill_pids "$(cat .frontend.pid 2>/dev/null || true)" && rm -f .frontend.pid

    # Also try to free ports 8081,8080,3000
    for p in 8081 8080 3000; do
        if command -v lsof >/dev/null 2>&1; then
            pids=$(lsof -ti :"$p" 2>/dev/null || true)
            if [ -n "$pids" ]; then
                echo "   Force freeing port $p (PIDs: $pids)"
                safe_kill_pids "$pids"
            fi
        fi
    done
    echo ""
fi

# ---------------------------
# Docker readiness (best-effort)
# ---------------------------
docker_ready=false
if command -v docker >/dev/null 2>&1 && docker info &>/dev/null && docker ps &>/dev/null; then
    docker_ready=true
fi

if [ "$docker_ready" != true ]; then
    echo "⚠️  Docker is not running. Attempting to start Docker..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        if [ -d "/Applications/Docker.app" ]; then
            if pgrep -f "Docker Desktop" >/dev/null 2>&1 || pgrep -f "com.docker.backend" >/dev/null 2>&1; then
                echo "   Docker Desktop process detected; waiting for daemon..."
            else
                echo "   Starting Docker Desktop..."
                open -a Docker 2>/dev/null || osascript -e 'tell application "Docker" to activate' 2>/dev/null || true
            fi
        else
            echo "   ❌ Docker Desktop not found in /Applications. Install from https://www.docker.com/products/docker-desktop"
            echo "   Exiting."
            exit 1
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if command -v systemctl >/dev/null 2>&1; then
            echo "   Starting docker via systemctl..."
            sudo systemctl start docker || { echo "   ❌ Failed to start docker via systemctl"; exit 1; }
        elif command -v service >/dev/null 2>&1; then
            echo "   Starting docker via service..."
            sudo service docker start || { echo "   ❌ Failed to start docker via service"; exit 1; }
        else
            echo "   ❌ Cannot auto-start Docker on this Linux. Start it manually and re-run."
            exit 1
        fi
    else
        echo "   ❌ Unsupported OS: $OSTYPE. Start Docker manually."
        exit 1
    fi

    # wait for daemon
    echo "   Waiting for Docker daemon to respond..."
    max_wait=60
    waited=0
    while [ $waited -lt $max_wait ]; do
        if docker info &>/dev/null && docker ps &>/dev/null; then
            docker_ready=true
            break
        fi
        sleep 2
        waited=$((waited + 2))
        if [ $((waited % 10)) -eq 0 ]; then
            printf "."
        fi
    done
    echo ""
    if [ "$docker_ready" = true ]; then
        echo "   ✅ Docker is running (waited ${waited}s)"
    else
        echo "   ⚠️  Docker daemon not ready after ${waited}s. Continuing (some parts may fail)."
    fi
else
    echo "✅ Docker appears to be running"
fi

echo ""

# ---------------------------
# Start etcd (via docker compose)
# ---------------------------
echo "1️⃣ Starting etcd..."
ETCD_HOST_PORT=2389
ETCD_CONTAINER_PORT=2379

if ! command -v docker >/dev/null 2>&1 || ! docker info &>/dev/null; then
    echo "   ❌ Docker not available - cannot start etcd"
    echo "   💡 Start Docker and run: $DOCKER_COMPOSE_CMD up -d etcd"
    exit 1
fi

# Check if container exists (running or stopped)
etcd_container_exists=false
if docker ps -a --format '{{.Names}}' 2>/dev/null | grep -q "^sharding-etcd$"; then
    etcd_container_exists=true
fi

# Check if container is running
etcd_running=false
if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^sharding-etcd$"; then
    etcd_running=true
    # Check if port is accessible (reliable check without needing shell access)
    if command -v nc >/dev/null 2>&1 && nc -z localhost "${ETCD_HOST_PORT}" 2>/dev/null; then
        echo "   ✅ etcd container is running and port ${ETCD_HOST_PORT} is accessible"
    else
        echo "   ⚠️  etcd container running but port ${ETCD_HOST_PORT} not accessible yet"
        # Container might be starting up, continue to wait loop
    fi
fi

# Start etcd if not running
if [ "$etcd_running" != true ]; then
    if [ "$etcd_container_exists" = true ]; then
        # Container exists but not running - try to start it
        echo "   Starting existing etcd container..."
        if ! docker start sharding-etcd >/dev/null 2>&1; then
            # If start fails, try removing and recreating
            echo "   ⚠️  Failed to start existing container, removing and recreating..."
            docker rm -f sharding-etcd >/dev/null 2>&1 || true
            $DOCKER_COMPOSE_CMD up -d etcd >/dev/null 2>&1 || {
                echo "   ❌ Failed to recreate etcd container"
                echo "   💡 Check docker-compose.yml and ensure Docker is running"
                exit 1
            }
        fi
    else
        echo "   Creating and starting etcd container..."
        $DOCKER_COMPOSE_CMD up -d etcd >/dev/null 2>&1 || {
            echo "   ❌ Failed to start etcd container"
            echo "   💡 Check docker-compose.yml and ensure Docker is running"
            exit 1
        }
    fi
    # Give container a moment to start
    sleep 2
fi

# Wait for etcd to be ready (check port accessibility)
echo "   Waiting for etcd to be ready on port ${ETCD_HOST_PORT}..."
max_wait=45
waited=0
etcd_ready=false

while [ $waited -lt $max_wait ]; do
    # Check if container is running
    if ! docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^sharding-etcd$"; then
        echo "   ❌ etcd container stopped unexpectedly"
        docker logs --tail 20 sharding-etcd 2>/dev/null || true
        exit 1
    fi
    
    # Check port accessibility (host port 2389 maps to container port 2379)
    if command -v nc >/dev/null 2>&1; then
        if nc -z localhost "${ETCD_HOST_PORT}" 2>/dev/null; then
            # Port is open, try a simple HTTP check
            if command -v curl >/dev/null 2>&1; then
                if curl -s -f -m 2 "http://localhost:${ETCD_HOST_PORT}/health" >/dev/null 2>&1 || \
                   curl -s -f -m 2 "http://localhost:${ETCD_HOST_PORT}/v3/health" >/dev/null 2>&1; then
                    etcd_ready=true
                    break
                fi
            else
                # Just port check is enough if curl not available
                etcd_ready=true
                break
            fi
        fi
    elif command -v timeout >/dev/null 2>&1; then
        # Fallback: try to connect with timeout
        if timeout 1 bash -c "echo > /dev/tcp/localhost/${ETCD_HOST_PORT}" 2>/dev/null; then
            etcd_ready=true
            break
        fi
    else
        # Last resort: just check if container is running for a few seconds
        if [ $waited -gt 5 ]; then
            etcd_ready=true
            break
        fi
    fi
    
    sleep 1
    waited=$((waited + 1))
    if [ $((waited % 5)) -eq 0 ]; then
        printf "."
    fi
done

echo ""
if [ "$etcd_ready" = true ]; then
    echo "   ✅ etcd is ready on port ${ETCD_HOST_PORT}"
    # Give etcd a moment to fully initialize
    sleep 2
else
    echo "   ⚠️  etcd may not be fully ready after ${waited}s"
    echo "   💡 Container is running, but health check timed out"
    echo "   💡 Check logs: docker logs sharding-etcd"
    echo "   ⚠️  Manager may fail to connect - waiting 5 more seconds..."
    sleep 5
fi

echo ""

# ---------------------------
# Build backend services (always rebuild on force-restart)
# ---------------------------
echo "2️⃣ Building backend services..."
mkdir -p bin
if command -v go >/dev/null 2>&1; then
    if [ "$FORCE_RESTART" = true ]; then
        echo "   Force-rebuild: cleaning previous binaries..."
        rm -f bin/manager bin/router || true
    fi
    go build -o bin/manager ./cmd/manager || { echo "   ❌ go build manager failed"; exit 1; }
    go build -o bin/router  ./cmd/router  || { echo "   ❌ go build router failed"; exit 1; }
    echo "   ✅ Build complete"
else
    echo "   ❌ Go not found in PATH; please install Go and re-run"
    exit 1
fi

echo ""

# ---------------------------
# Start manager
# ---------------------------
echo "3️⃣ Starting manager..."
MANAGER_PORT=8081
MANAGER_PID_FILE=".manager.pid"

# If force restart, ensure port and pids are freed
if [ "$FORCE_RESTART" = true ]; then
    echo "   --force-restart: ensure manager stopped"
    if [ -f "$MANAGER_PID_FILE" ]; then
        safe_kill_pids "$(cat "$MANAGER_PID_FILE" 2>/dev/null || true)"
        rm -f "$MANAGER_PID_FILE"
    fi
    if command -v lsof >/dev/null 2>&1; then
        pids=$(lsof -ti :"$MANAGER_PORT" 2>/dev/null || true)
        safe_kill_pids "$pids"
    fi
fi

# Remove stale pid file if process dead
if [ -f "$MANAGER_PID_FILE" ]; then
    old_pid=$(cat "$MANAGER_PID_FILE" 2>/dev/null || echo "")
    if [ -n "$old_pid" ] && ps -p "$old_pid" >/dev/null 2>&1; then
        echo "   Manager already running (PID: $old_pid)"
    else
        rm -f "$MANAGER_PID_FILE"
    fi
fi

if [ ! -f "$MANAGER_PID_FILE" ]; then
    export CONFIG_PATH="${CONFIG_PATH:-configs/manager.json}"
    export JWT_SECRET="${JWT_SECRET:-development-secret-not-for-production-use-min-32-chars-please-change}"

    if [ -f .env ]; then
        echo "   📝 Loading environment from .env"
        set -a
        # shellcheck disable=SC1091
        source .env
        set +a
    fi

    echo "   🚀 Starting manager (logs -> logs/manager.log)..."
    ./bin/manager > logs/manager.log 2>&1 &
    manager_pid=$!
    echo "$manager_pid" > "$MANAGER_PID_FILE"

    sleep 2
    if ! ps -p "$manager_pid" >/dev/null 2>&1; then
        echo "   ❌ Manager process died. Tail logs/manager.log"
        tail -n 40 logs/manager.log || true
        exit 1
    fi

    echo "   Waiting for manager to respond on http://localhost:${MANAGER_PORT}/api/v1/health ..."
    max_wait=60
    elapsed=0
    http_ok=false
    while [ $elapsed -lt $max_wait ]; do
        if lsof -ti :"$MANAGER_PORT" >/dev/null 2>&1 || curl -s -f -m 2 "http://localhost:${MANAGER_PORT}/api/v1/health" >/dev/null 2>&1; then
            if curl -s -f -m 2 "http://localhost:${MANAGER_PORT}/api/v1/health" >/dev/null 2>&1; then
                http_ok=true
                break
            fi
        fi
        if ! ps -p "$manager_pid" >/dev/null 2>&1; then
            echo "   ❌ Manager process exited during startup. Check logs."
            tail -n 40 logs/manager.log || true
            exit 1
        fi
        sleep 2
        elapsed=$((elapsed + 2))
        if [ $((elapsed % 10)) -eq 0 ]; then
            printf "."
        fi
    done
    echo ""
    if [ "$http_ok" = true ]; then
        echo "   ✅ Manager started successfully (PID: $manager_pid)"
    else
        echo "   ⚠️  Manager running (PID: $manager_pid) but health endpoint not responding yet. Check logs/manager.log"
    fi
else
    echo "   ℹ️  Manager already running"
fi

echo ""

# ---------------------------
# Start router
# ---------------------------
echo "4️⃣ Starting router..."
ROUTER_PORT=8080
ROUTER_PID_FILE=".router.pid"

if [ "$FORCE_RESTART" = true ]; then
    echo "   --force-restart: ensure router stopped"
    [ -f "$ROUTER_PID_FILE" ] && safe_kill_pids "$(cat "$ROUTER_PID_FILE" 2>/dev/null || true)" && rm -f "$ROUTER_PID_FILE"
    if command -v lsof >/dev/null 2>&1; then
        pids=$(lsof -ti :"$ROUTER_PORT" 2>/dev/null || true)
        safe_kill_pids "$pids"
    fi
fi

if [ -f "$ROUTER_PID_FILE" ]; then
    old_router_pid=$(cat "$ROUTER_PID_FILE" 2>/dev/null || echo "")
    if [ -n "$old_router_pid" ] && ps -p "$old_router_pid" >/dev/null 2>&1; then
        echo "   Router already running (PID: $old_router_pid)"
    else
        rm -f "$ROUTER_PID_FILE"
    fi
fi

if ! lsof -ti :"$ROUTER_PORT" >/dev/null 2>&1; then
    echo "   🚀 Starting router (logs -> logs/router.log)..."
    ./bin/router > logs/router.log 2>&1 &
    router_pid=$!
    echo "$router_pid" > "$ROUTER_PID_FILE"
    sleep 2
    if ps -p "$router_pid" >/dev/null 2>&1; then
        echo "   ✅ Router started (PID: $router_pid)"
    else
        echo "   ⚠️  Router failed to start - check logs/router.log"
    fi
else
    echo "   ⚠️  Router port ${ROUTER_PORT} still in use; skipping router start"
fi

echo ""

# ---------------------------
# Start frontend
# ---------------------------
echo "5️⃣ Starting frontend..."
FRONTEND_PORT=3000
FRONTEND_PID_FILE=".frontend.pid"

if [ "$FORCE_RESTART" = true ]; then
    echo "   --force-restart: ensure frontend stopped"
    [ -f "$FRONTEND_PID_FILE" ] && safe_kill_pids "$(cat "$FRONTEND_PID_FILE" 2>/dev/null || true)" && rm -f "$FRONTEND_PID_FILE"
    if command -v lsof >/dev/null 2>&1; then
        pids=$(lsof -ti :"$FRONTEND_PORT" 2>/dev/null || true)
        safe_kill_pids "$pids"
    fi
fi

if [ -f "$FRONTEND_PID_FILE" ]; then
    old_front_pid=$(cat "$FRONTEND_PID_FILE" 2>/dev/null || echo "")
    if [ -n "$old_front_pid" ] && ps -p "$old_front_pid" >/dev/null 2>&1; then
        echo "   Frontend appears to be already running (PID: $old_front_pid)"
    else
        rm -f "$FRONTEND_PID_FILE"
    fi
fi

if ! lsof -ti :"$FRONTEND_PORT" >/dev/null 2>&1; then
    if [ -d ui ]; then
        echo "   Starting frontend from ./ui (logs -> logs/frontend.log)..."
        pushd ui >/dev/null
        if [ ! -d "node_modules" ]; then
            echo "   Installing frontend dependencies (npm install)..."
            npm install
        fi
        npm run dev > ../logs/frontend.log 2>&1 &
        front_pid=$!
        popd >/dev/null
        echo "$front_pid" > "$FRONTEND_PID_FILE"
        sleep 3
        if ps -p "$front_pid" >/dev/null 2>&1; then
            echo "   ✅ Frontend started (PID: $front_pid)"
        else
            echo "   ❌ Frontend failed to start. Check logs/frontend.log"
            tail -n 40 logs/frontend.log || true
        fi
    else
        echo "   ❌ ./ui directory not found; cannot start frontend"
    fi
else
    echo "   ⚠️  Port ${FRONTEND_PORT} is in use; skipping frontend start"
fi

echo ""
echo "✅ Start sequence finished (force-restart=${FORCE_RESTART}). Some components may still be initializing."
echo ""
echo "📊 Services (expected):"
echo "   Manager:  http://localhost:${MANAGER_PORT}"
echo "   Router:   http://localhost:${ROUTER_PORT}"
echo "   Frontend: http://localhost:${FRONTEND_PORT}"
echo "   etcd:     http://localhost:2379"
echo ""
echo "📝 Logs:"
echo "   Manager:  tail -f logs/manager.log"
echo "   Router:   tail -f logs/router.log"
echo "   Frontend: tail -f logs/frontend.log"
echo ""
echo "🛑 To stop: scripts/stop.sh"
echo ""
