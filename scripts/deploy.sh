#!/bin/bash
# deploy.sh - Zero-downtime deployment script for TalentSynapse
# This script is executed on the VPS by GitHub Actions
set -e

APP_DIR="/var/www/myapp"
SERVICE_NAME="myapp"
BINARY_NAME="server"

cd "$APP_DIR"

echo "🚀 Starting deployment..."

# Check if new binary exists
if [ ! -f "server.new" ]; then
    echo "❌ Error: server.new not found"
    exit 1
fi

# Make executable
chmod +x server.new

# Backup current binary
if [ -f "$BINARY_NAME" ]; then
    echo "📦 Backing up current binary..."
    cp "$BINARY_NAME" "${BINARY_NAME}.backup"
fi

# Atomic swap
echo "🔄 Swapping binaries..."
mv server.new "$BINARY_NAME"

# Run database migrations (optional - uncomment if needed)
# echo "🗃️ Running database migrations..."
# /var/www/myapp/goose -dir ./migrations postgres "$DATABASE_URL" up

# Restart service
echo "🔄 Restarting service..."
if systemctl restart "$SERVICE_NAME" 2>/dev/null; then
    echo "✅ Restarted without sudo"
elif sudo systemctl restart "$SERVICE_NAME" 2>/dev/null; then
    echo "✅ Restarted with sudo"
else
    echo "❌ Failed to restart service"
    echo "⚠️  Please configure passwordless sudo for systemctl"
    exit 1
fi

# Wait for service to start
sleep 3

# Check if service is running
if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null || sudo systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "✅ Service restarted successfully!"

    # Clean up old backup (keep only 1)
    if [ -f "${BINARY_NAME}.backup" ]; then
        rm -f "${BINARY_NAME}.old"
        mv "${BINARY_NAME}.backup" "${BINARY_NAME}.old"
    fi
else
    echo "❌ Service failed to start! Rolling back..."

    # Rollback
    if [ -f "${BINARY_NAME}.backup" ]; then
        echo "🔙 Rolling back to previous version..."
        mv "${BINARY_NAME}.backup" "$BINARY_NAME"
        systemctl restart "$SERVICE_NAME" 2>/dev/null || sudo systemctl restart "$SERVICE_NAME"
        echo "🔙 Rolled back to previous version"
    fi

    exit 1
fi

echo "🎉 Deployment complete!"
