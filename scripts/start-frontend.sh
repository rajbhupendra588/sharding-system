#!/bin/bash

set -e

# Get the project root directory (parent of scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "🎨 Starting Sharding System Frontend UI..."

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Error: Node.js is not installed or not in PATH"
    echo "Please install Node.js 18+ from https://nodejs.org"
    exit 1
fi

# Check Node.js version
NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    echo "❌ Error: Node.js 18+ is required. Current version: $(node -v)"
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "❌ Error: npm is not installed"
    exit 1
fi

cd ui

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo "📦 Installing frontend dependencies..."
    npm install
else
    echo "✅ Dependencies already installed"
fi

# Start dev server
echo "🚀 Starting frontend development server..."
echo ""
echo "✅ Frontend UI will be available at: http://localhost:3000"
echo ""
echo "🛑 Press Ctrl+C to stop the server"
echo ""

npm run dev

