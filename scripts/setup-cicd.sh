#!/bin/bash
# setup-cicd.sh - Quick setup script for GitHub Actions CI/CD
# Run this on your VPS to prepare for automated deployments

set -e

APP_DIR="/var/www/myapp"
SERVICE_NAME="myapp"
DEPLOY_SCRIPT="$APP_DIR/deploy.sh"

echo "🔧 Setting up CI/CD deployment on VPS..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
   echo "❌ Please run as root: sudo bash setup-cicd.sh"
   exit 1
fi

# Ensure app directory exists
if [ ! -d "$APP_DIR" ]; then
    echo "📁 Creating $APP_DIR..."
    mkdir -p "$APP_DIR"
fi

# Set proper ownership
echo "🔐 Setting ownership..."
chown -R www-data:www-data "$APP_DIR"
chmod -R 755 "$APP_DIR"

# Check if deploy script exists
if [ ! -f "$DEPLOY_SCRIPT" ]; then
    echo "⚠️  Deploy script not found at $DEPLOY_SCRIPT"
    echo "📥 You'll need to upload it manually:"
    echo "   scp scripts/deploy.sh root@talentsynapse:$DEPLOY_SCRIPT"
else
    echo "✅ Deploy script found"
    chmod +x "$DEPLOY_SCRIPT"
fi

# Check service file
if [ ! -f "/etc/systemd/system/$SERVICE_NAME.service" ]; then
    echo "❌ Service file not found at /etc/systemd/system/$SERVICE_NAME.service"
    exit 1
else
    echo "✅ Service file found"
fi

# Reload systemd
echo "🔄 Reloading systemd..."
systemctl daemon-reload

# Check if service is enabled
if systemctl is-enabled --quiet "$SERVICE_NAME"; then
    echo "✅ Service is enabled"
else
    echo "📋 Enabling service..."
    systemctl enable "$SERVICE_NAME"
fi

# Test current service status
if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "✅ Service is currently running"
else
    echo "⚠️  Service is not running"
    echo "   Start it with: sudo systemctl start $SERVICE_NAME"
fi

echo ""
echo "✅ VPS setup complete!"
echo ""
echo "📋 Next steps:"
echo "1. Ensure deploy script is uploaded:"
echo "   scp scripts/deploy.sh root@talentsynapse:$DEPLOY_SCRIPT"
echo ""
echo "2. Get your private SSH key for GitHub:"
echo "   cat ~/.ssh/id_ed25519"
echo ""
echo "3. Add GitHub Secrets (see docs/GITHUB_ACTIONS_SETUP.md):"
echo "   - VPS_HOST: talentsynapse (or IP)"
echo "   - VPS_USER: root"
echo "   - VPS_SSH_KEY: (output from step 2)"
echo "   - VPS_DOMAIN: talentsynapse.org"
echo ""
echo "4. Test deployment:"
echo "   git commit -m 'test: CI/CD' && git push origin main"
