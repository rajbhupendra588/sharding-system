#!/bin/bash

# Production Setup Script
# This script helps set up the environment for production deployment

set -e

echo "🔒 Production Setup Script"
echo "=========================="
echo ""

# Check if JWT_SECRET is set
if [ -z "$JWT_SECRET" ]; then
    echo "⚠️  JWT_SECRET is not set!"
    echo "Generating a secure JWT secret..."
    export JWT_SECRET=$(openssl rand -base64 32)
    echo "✅ Generated JWT_SECRET: $JWT_SECRET"
    echo ""
    echo "⚠️  IMPORTANT: Save this JWT_SECRET securely!"
    echo "Add it to your environment or secrets management system."
    echo ""
else
    if [ ${#JWT_SECRET} -lt 32 ]; then
        echo "❌ ERROR: JWT_SECRET must be at least 32 characters"
        echo "Current length: ${#JWT_SECRET}"
        exit 1
    fi
    echo "✅ JWT_SECRET is set (length: ${#JWT_SECRET})"
fi

# Check CORS configuration
if [ -z "$CORS_ALLOWED_ORIGINS" ]; then
    echo "⚠️  CORS_ALLOWED_ORIGINS is not set"
    echo "Using default: * (allows all origins)"
    echo "⚠️  WARNING: Restrict this in production!"
    export CORS_ALLOWED_ORIGINS="*"
else
    echo "✅ CORS_ALLOWED_ORIGINS is set: $CORS_ALLOWED_ORIGINS"
fi

# Validate configuration files
echo ""
echo "📋 Validating configuration files..."

if [ ! -f "configs/manager.json" ]; then
    echo "❌ ERROR: configs/manager.json not found"
    exit 1
fi

if [ ! -f "configs/router.json" ]; then
    echo "❌ ERROR: configs/router.json not found"
    exit 1
fi

# Check if RBAC is enabled
if grep -q '"enable_rbac":\s*true' configs/manager.json; then
    echo "✅ RBAC is enabled in manager config"
else
    echo "⚠️  WARNING: RBAC is not enabled in manager config"
    echo "   Set 'enable_rbac: true' in configs/manager.json"
fi

if grep -q '"enable_rbac":\s*true' configs/router.json; then
    echo "✅ RBAC is enabled in router config"
else
    echo "⚠️  WARNING: RBAC is not enabled in router config"
    echo "   Set 'enable_rbac: true' in configs/router.json"
fi

echo ""
echo "✅ Production setup validation complete!"
echo ""
echo "📝 Next steps:"
echo "   1. Ensure JWT_SECRET is set in your deployment environment"
echo "   2. Set CORS_ALLOWED_ORIGINS to your production domains"
echo "   3. Enable TLS in production (set enable_tls: true)"
echo "   4. Review and update security settings in config files"
echo "   5. Set up monitoring and alerting"
echo ""


