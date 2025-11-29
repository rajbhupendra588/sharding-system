#!/bin/bash
# Quick start script with GitHub OAuth

cd "$(dirname "$0")"

echo "🚀 Starting Sharding System with GitHub OAuth..."

# Load .env file if it exists
if [ -f .env ]; then
    echo "📝 Loading .env file..."
    set -a
    source .env
    set +a
fi

# Set GitHub OAuth if not already set
export GITHUB_OAUTH_CLIENT_ID="${GITHUB_OAUTH_CLIENT_ID:-Ov23li8EHKaytnZ8o0Wx}"
export GITHUB_OAUTH_CLIENT_SECRET="${GITHUB_OAUTH_CLIENT_SECRET:-789dbcc19fac32e3d816d64280b492b84cbea8d8}"
export BASE_URL="${BASE_URL:-http://localhost:8081}"

echo "✅ GitHub OAuth configured"
echo "   Client ID: ${GITHUB_OAUTH_CLIENT_ID:0:10}..."
echo ""

# Check if Docker is running
if ! docker info &> /dev/null; then
    echo "⚠️  Docker is not running. Please start Docker Desktop first."
    exit 1
fi

# Start etcd
echo "📦 Starting etcd..."
docker-compose up -d etcd
sleep 3

# Start manager
echo "🔧 Starting manager server..."
lsof -ti :8081 | xargs kill 2>/dev/null || true
sleep 1

./bin/manager > logs/manager.log 2>&1 &
MANAGER_PID=$!
echo $MANAGER_PID > .manager.pid
sleep 3

if ps -p $MANAGER_PID > /dev/null 2>&1; then
    echo "✅ Manager started (PID: $MANAGER_PID)"
    echo ""
    echo "📊 Check logs: tail -f logs/manager.log | grep -i oauth"
    echo "🌐 Frontend: http://localhost:3000/login"
    echo "🔗 API: http://localhost:8081/api/v1/auth/oauth/providers"
else
    echo "❌ Manager failed to start. Check logs/manager.log"
    exit 1
fi
