# Production Migration Guide - Skillsphere PWA

Complete guide to deploying Skillsphere PWA to production on a VPS (Hetzner or Digital Ocean) with PostgreSQL, QUIC/HTTP3, systemd, and custom domain.

---

## Table of Contents

1. [VPS Setup](#1-vps-setup)
2. [Database Setup with Docker](#2-database-setup-with-docker)
3. [Application Deployment](#3-application-deployment)
4. [QUIC/HTTP3 with Caddy](#4-quichttp3-with-caddy)
5. [Systemd Service Configuration](#5-systemd-service-configuration)
6. [Domain Configuration](#6-domain-configuration)
7. [Database Migrations](#7-database-migrations)
8. [CI/CD Pipeline](#8-cicd-pipeline)
9. [Monitoring & Maintenance](#9-monitoring--maintenance)
10. [Rollback Procedures](#10-rollback-procedures)

---

## 1. VPS Setup

### 1.1 Choose Your VPS Provider

**Hetzner (Recommended for EU):**
- Best price/performance ratio
- €4.15/month for 2 vCPU, 4GB RAM
- Data centers in Germany, Finland, USA

**Digital Ocean:**
- $12/month for 2 vCPU, 4GB RAM
- Global data centers
- Better for non-EU locations

### 1.2 Initial VPS Configuration

```bash
# SSH into your VPS
ssh root@your-vps-ip

# Update system
apt update && apt upgrade -y

# Install essential packages
apt install -y \
  curl \
  git \
  wget \
  ufw \
  fail2ban \
  htop \
  vim \
  docker.io \
  docker-compose

# Create non-root user
adduser skillsphere
usermod -aG sudo skillsphere
usermod -aG docker skillsphere

# Enable Docker on boot
systemctl enable docker
systemctl start docker

# Configure firewall
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS
ufw allow 443/udp   # QUIC/HTTP3
ufw enable

# Configure fail2ban for SSH protection
systemctl enable fail2ban
systemctl start fail2ban

# Switch to non-root user
su - skillsphere
```

### 1.3 SSH Key Setup (Local Machine)

```bash
# Generate SSH key if you don't have one
ssh-keygen -t ed25519 -C "your-email@example.com"

# Copy SSH key to VPS
ssh-copy-id skillsphere@your-vps-ip

# Test passwordless login
ssh skillsphere@your-vps-ip
```

---

## 2. Database Setup with Docker

### 2.1 Create Production Docker Compose File

On your VPS, create `/home/skillsphere/database/docker-compose.yml`:

```yaml
version: '3.8'

services:
  db:
    image: postgres:17-alpine
    container_name: skillsphere_postgres
    environment:
      POSTGRES_USER: skillsphere
      POSTGRES_PASSWORD: ${DB_PASSWORD}  # Set via .env file
      POSTGRES_DB: skillsphere_prod
      PGDATA: /var/lib/postgresql/data/pgdata
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./config/postgresql.conf:/etc/postgresql/postgresql.conf:ro
      - ./backups:/backups
    ports:
      - "127.0.0.1:5432:5432"  # Only localhost access
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U skillsphere"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped
    networks:
      - skillsphere_network
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

volumes:
  postgres_data:
    driver: local

networks:
  skillsphere_network:
    driver: bridge
```

### 2.2 Create PostgreSQL Configuration

Create `/home/skillsphere/database/config/postgresql.conf`:

```ini
# Connection Settings
max_connections = 100
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB

# WAL Settings (for backups)
wal_level = replica
max_wal_size = 1GB
min_wal_size = 80MB

# Performance
random_page_cost = 1.1
effective_io_concurrency = 200

# Logging
logging_collector = on
log_directory = 'pg_log'
log_filename = 'postgresql-%Y-%m-%d.log'
log_rotation_age = 1d
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
log_min_duration_statement = 1000  # Log slow queries (>1s)

# Locale
lc_messages = 'en_US.utf8'
lc_monetary = 'en_US.utf8'
lc_numeric = 'en_US.utf8'
lc_time = 'en_US.utf8'
```

### 2.3 Environment Configuration

Create `/home/skillsphere/database/.env`:

```bash
# IMPORTANT: Generate a strong password
DB_PASSWORD=your-super-strong-password-here
```

**Generate a strong password:**
```bash
openssl rand -base64 32
```

### 2.4 Start Database

```bash
cd /home/skillsphere/database
docker-compose up -d

# Check logs
docker-compose logs -f

# Verify database is running
docker-compose ps
```

### 2.5 Create Database User

```bash
# Connect to database
docker exec -it skillsphere_postgres psql -U skillsphere -d skillsphere_prod

# Inside PostgreSQL shell:
-- Verify connection
\conninfo

-- Create extensions (if needed)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- For full-text search

-- Exit
\q
```

---

## 3. Application Deployment

### 3.1 Create Application Directory

```bash
sudo mkdir -p /var/www/skillsphere
sudo chown -R skillsphere:skillsphere /var/www/skillsphere
cd /var/www/skillsphere
```

### 3.2 Build Binary Locally

```bash
# On your local machine
cd /Users/fernando_idwell/Projects/Skillsphere/skillsphere/skillsphere-pwa

# Build production binary
make build-prod
# Or: ./scripts/build-prod.sh

# Verify binary works
./bin/server --version
```

### 3.3 Deploy to VPS

```bash
# Copy binary to VPS
scp bin/server skillsphere@your-vps-ip:/var/www/skillsphere/

# Copy assets (if needed)
scp -r assets skillsphere@your-vps-ip:/var/www/skillsphere/

# Make binary executable
ssh skillsphere@your-vps-ip "chmod +x /var/www/skillsphere/server"
```

### 3.4 Environment Configuration

Create `/var/www/skillsphere/.env`:

```bash
# Application
GO_ENV=production
PORT=8080

# Database
DATABASE_URL=postgres://skillsphere:your-db-password@127.0.0.1:5432/skillsphere_prod?sslmode=disable

# Session
SESSION_SECRET=$(openssl rand -base64 32)

# Admin (change these!)
ADMIN_EMAIL=admin@yourdomain.com
ADMIN_PASSWORD=your-secure-admin-password

# Optional: Feature flags
ENABLE_REGISTRATION=true
ENABLE_ANALYTICS=false
```

---

## 4. QUIC/HTTP3 with Caddy

### 4.1 Install Caddy

```bash
# Install Caddy (Debian/Ubuntu)
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl

curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | \
  sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg

curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | \
  sudo tee /etc/apt/sources.list.d/caddy-stable.list

sudo apt update
sudo apt install caddy

# Verify installation
caddy version
```

### 4.2 Configure Caddy with QUIC/HTTP3

Create `/etc/caddy/Caddyfile`:

```caddyfile
{
    email admin@yourdomain.com  # For Let's Encrypt notifications

    # Enable HTTP/3 (QUIC) globally
    servers {
        protocols h1 h2 h3
    }
}

yourdomain.com {
    # Enable HTTP/3 for this domain
    protocols h1 h2 h3

    # Compression
    encode gzip zstd

    # Security headers
    header {
        # HSTS
        Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"

        # Security
        X-Content-Type-Options "nosniff"
        X-Frame-Options "SAMEORIGIN"
        X-XSS-Protection "1; mode=block"
        Referrer-Policy "strict-origin-when-cross-origin"

        # CSP (adjust based on your needs)
        Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'"

        # Remove Server header
        -Server
    }

    # Cache static assets aggressively
    @static {
        path /assets/css/* /assets/js/* /assets/fonts/* /assets/images/*
        not path /assets/static/sw.js
    }
    header @static {
        Cache-Control "public, max-age=31536000, immutable"
    }

    # Service Worker - never cache
    header /assets/static/sw.js {
        Cache-Control "no-cache, no-store, must-revalidate"
        Pragma "no-cache"
        Expires "0"
    }

    # Manifest - cache for 1 day
    header /assets/static/manifest.json {
        Cache-Control "public, max-age=86400"
    }

    # Reverse proxy to Go application
    reverse_proxy localhost:8080 {
        # Health checks
        health_uri /health
        health_interval 10s
        health_timeout 5s

        # Retry on failure
        fail_duration 30s
        max_fails 3
        unhealthy_status 500 502 503

        # Forward real IP
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-Proto {scheme}
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Host {host}
    }

    # Custom error pages (optional)
    handle_errors {
        @maintenance expression {http.error.status_code} == 503
        rewrite @maintenance /maintenance.html
        file_server
    }

    # Logging
    log {
        output file /var/log/caddy/skillsphere.log {
            roll_size 10mb
            roll_keep 5
        }
        format json
        level INFO
    }
}

# Redirect www to non-www
www.yourdomain.com {
    redir https://yourdomain.com{uri} permanent
}
```

### 4.3 Test and Start Caddy

```bash
# Test configuration
sudo caddy validate --config /etc/caddy/Caddyfile

# Format Caddyfile
sudo caddy fmt --overwrite /etc/caddy/Caddyfile

# Reload Caddy
sudo systemctl reload caddy

# Check status
sudo systemctl status caddy

# View logs
sudo journalctl -u caddy -f
```

### 4.4 Verify HTTP/3 is Working

```bash
# Using curl (requires curl with HTTP/3 support)
curl --http3 -I https://yourdomain.com

# Check in browser DevTools:
# Network tab → Protocol column should show "h3" or "http/3"
```

---

## 5. Systemd Service Configuration

### 5.1 Create Systemd Service

Create `/etc/systemd/system/skillsphere.service`:

```ini
[Unit]
Description=Skillsphere PWA - Go Backend
Documentation=https://github.com/yourusername/skillsphere
After=network-online.target docker.service
Wants=network-online.target
Requires=docker.service

[Service]
Type=simple
User=skillsphere
Group=skillsphere
WorkingDirectory=/var/www/skillsphere

# Environment file
EnvironmentFile=/var/www/skillsphere/.env

# The binary (Caddy handles TLS, so no TLS in Go)
ExecStart=/var/www/skillsphere/server

# Restart policy
Restart=always
RestartSec=10
StartLimitInterval=400
StartLimitBurst=3

# Graceful shutdown
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/www/skillsphere
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
RestrictRealtime=true
RestrictNamespaces=true
LockPersonality=true

# Resource limits
LimitNOFILE=65535
LimitNPROC=512

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=skillsphere

[Install]
WantedBy=multi-user.target
```

### 5.2 Enable and Start Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service (start on boot)
sudo systemctl enable skillsphere

# Start service
sudo systemctl start skillsphere

# Check status
sudo systemctl status skillsphere

# View logs
sudo journalctl -u skillsphere -f

# Tail last 100 lines
sudo journalctl -u skillsphere -n 100 --no-pager
```

### 5.3 Service Management Commands

```bash
# Start
sudo systemctl start skillsphere

# Stop
sudo systemctl stop skillsphere

# Restart
sudo systemctl restart skillsphere

# Reload (if supported)
sudo systemctl reload skillsphere

# Status
sudo systemctl status skillsphere

# Enable (start on boot)
sudo systemctl enable skillsphere

# Disable (don't start on boot)
sudo systemctl disable skillsphere

# View logs (live)
sudo journalctl -u skillsphere -f

# View logs (last 100 lines)
sudo journalctl -u skillsphere -n 100
```

---

## 6. Domain Configuration

### 6.1 Purchase Domain

**Recommended registrars:**
- Namecheap
- Cloudflare Registrar (best pricing)
- Google Domains / Squarespace
- Porkbun

### 6.2 Configure DNS

Add these DNS records at your registrar:

```
Type    Name    Value               TTL
A       @       your-vps-ip         300
A       www     your-vps-ip         300
AAAA    @       your-vps-ipv6       300  (if available)
```

**Example:**
```
A       @       95.217.123.45       300
A       www     95.217.123.45       300
```

### 6.3 Verify DNS Propagation

```bash
# Check DNS
dig yourdomain.com +short

# Check from different location
nslookup yourdomain.com 8.8.8.8

# Check propagation globally
# Visit: https://www.whatsmydns.net/
```

### 6.4 Wait for SSL Certificate

```bash
# Caddy will automatically obtain Let's Encrypt certificate
# Watch logs for certificate acquisition
sudo journalctl -u caddy -f

# You should see:
# "certificate obtained successfully"
```

### 6.5 Test Your Domain

```bash
# Test HTTP → HTTPS redirect
curl -I http://yourdomain.com
# Should return: 301 or 308 redirect to https://

# Test HTTPS
curl -I https://yourdomain.com
# Should return: 200 OK

# Test SSL
curl -vI https://yourdomain.com 2>&1 | grep -E "SSL|TLS"
# Should show TLS 1.3

# Test HTTP/3
curl --http3 -I https://yourdomain.com
# Should work if curl supports HTTP/3
```

---

## 7. Database Migrations

### 7.1 Install Goose on VPS

```bash
# On VPS
go install github.com/pressly/goose/v3/cmd/goose@latest

# Or download binary
wget https://github.com/pressly/goose/releases/download/v3.17.0/goose_linux_amd64
sudo mv goose_linux_amd64 /usr/local/bin/goose
sudo chmod +x /usr/local/bin/goose

# Verify
goose --version
```

### 7.2 Copy Migration Files

```bash
# From local machine
scp -r internal/database/migrations skillsphere@your-vps-ip:/var/www/skillsphere/

# Verify on VPS
ssh skillsphere@your-vps-ip "ls -la /var/www/skillsphere/migrations"
```

### 7.3 Run Migrations

```bash
# SSH into VPS
ssh skillsphere@your-vps-ip

# Set database URL
export DATABASE_URL="postgres://skillsphere:your-db-password@127.0.0.1:5432/skillsphere_prod?sslmode=disable"

# Check migration status
goose -dir /var/www/skillsphere/migrations postgres "$DATABASE_URL" status

# Run migrations
goose -dir /var/www/skillsphere/migrations postgres "$DATABASE_URL" up

# Verify
goose -dir /var/www/skillsphere/migrations postgres "$DATABASE_URL" status
```

### 7.4 Create Migration Script

Create `/var/www/skillsphere/migrate.sh`:

```bash
#!/bin/bash
set -e

# Configuration
MIGRATIONS_DIR="/var/www/skillsphere/migrations"
DB_URL="postgres://skillsphere:$(grep DB_PASSWORD /home/skillsphere/database/.env | cut -d'=' -f2)@127.0.0.1:5432/skillsphere_prod?sslmode=disable"

# Check if goose is installed
if ! command -v goose &> /dev/null; then
    echo "Error: goose is not installed"
    exit 1
fi

# Run migrations
echo "Running database migrations..."
goose -dir "$MIGRATIONS_DIR" postgres "$DB_URL" up

echo "Migration complete!"
goose -dir "$MIGRATIONS_DIR" postgres "$DB_URL" status
```

Make it executable:

```bash
chmod +x /var/www/skillsphere/migrate.sh
```

### 7.5 Add to Deployment Script

When deploying, always run migrations first:

```bash
# 1. Stop application
sudo systemctl stop skillsphere

# 2. Backup database (optional but recommended)
/home/skillsphere/database/backup.sh

# 3. Run migrations
/var/www/skillsphere/migrate.sh

# 4. Deploy new binary
cp server.new server

# 5. Start application
sudo systemctl start skillsphere
```

---

## 8. CI/CD Pipeline

### 8.1 Update Makefile for CI/CD

Add to your Makefile (`/Users/fernando_idwell/Projects/Skillsphere/skillsphere/skillsphere-pwa/Makefile`):

```makefile
# =========================================================================
# CI/CD Deployment
# =========================================================================

.PHONY: ci-migrate deploy-prod

# Run migrations in CI/CD pipeline
ci-migrate: ## Run database migrations (CI/CD)
	@echo "🔄 Running database migrations..."
	goose -dir $(GOOSE_MIGRATION_DIR) $(GOOSE_DRIVER) "$(DB_DSN)" up
	@echo "✅ Migrations complete"

# Full production deployment
deploy-prod: build-prod ## Build and deploy to production
	@echo "🚀 Starting production deployment..."
	@echo "1️⃣ Building binary..."
	@./scripts/build-prod.sh
	@echo ""
	@echo "2️⃣ Ready to deploy. Run these commands on your VPS:"
	@echo ""
	@echo "  # Stop service"
	@echo "  sudo systemctl stop skillsphere"
	@echo ""
	@echo "  # Run migrations"
	@echo "  /var/www/skillsphere/migrate.sh"
	@echo ""
	@echo "  # Deploy new binary"
	@echo "  cp /var/www/skillsphere/server.new /var/www/skillsphere/server"
	@echo ""
	@echo "  # Start service"
	@echo "  sudo systemctl start skillsphere"
	@echo ""
	@echo "  # Check status"
	@echo "  sudo systemctl status skillsphere"
	@echo ""
```

### 8.2 GitHub Actions Workflow

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy to Production

on:
  push:
    branches:
      - main
  workflow_dispatch:  # Manual trigger

env:
  GO_VERSION: '1.22'

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Run tests
        run: make test

  build:
    name: Build
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Install dependencies
        run: make ci-setup

      - name: Build binary
        run: make build-prod

      - name: Upload binary
        uses: actions/upload-artifact@v4
        with:
          name: server-binary
          path: bin/server
          retention-days: 7

  deploy:
    name: Deploy to VPS
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'

    steps:
      - uses: actions/checkout@v4

      - name: Download binary
        uses: actions/download-artifact@v4
        with:
          name: server-binary
          path: bin/

      - name: Setup SSH
        run: |
          mkdir -p ~/.ssh
          echo "${{ secrets.VPS_SSH_KEY }}" > ~/.ssh/id_ed25519
          chmod 600 ~/.ssh/id_ed25519
          ssh-keyscan -H ${{ secrets.VPS_IP }} >> ~/.ssh/known_hosts

      - name: Copy migrations to VPS
        run: |
          scp -r internal/database/migrations \
            ${{ secrets.VPS_USER }}@${{ secrets.VPS_IP }}:/var/www/skillsphere/

      - name: Run database migrations
        run: |
          ssh ${{ secrets.VPS_USER }}@${{ secrets.VPS_IP }} \
            "/var/www/skillsphere/migrate.sh"

      - name: Deploy binary
        run: |
          # Copy new binary
          scp bin/server \
            ${{ secrets.VPS_USER }}@${{ secrets.VPS_IP }}:/var/www/skillsphere/server.new

          # Atomic swap and restart
          ssh ${{ secrets.VPS_USER }}@${{ secrets.VPS_IP }} << 'EOF'
            cd /var/www/skillsphere
            sudo systemctl stop skillsphere
            mv server server.old
            mv server.new server
            chmod +x server
            sudo systemctl start skillsphere
            sleep 5
            sudo systemctl status skillsphere
          EOF

      - name: Health check
        run: |
          sleep 10
          curl -f https://${{ secrets.DOMAIN }}/health || exit 1

      - name: Cleanup old binary
        if: success()
        run: |
          ssh ${{ secrets.VPS_USER }}@${{ secrets.VPS_IP }} \
            "rm -f /var/www/skillsphere/server.old"
```

### 8.3 Configure GitHub Secrets

In your GitHub repo settings → Secrets and variables → Actions:

```
VPS_IP=your-vps-ip
VPS_USER=skillsphere
VPS_SSH_KEY=<paste your private SSH key>
DOMAIN=yourdomain.com
```

### 8.4 Manual Deployment Script

Create `scripts/deploy.sh`:

```bash
#!/bin/bash
set -e

# Configuration
VPS_USER="skillsphere"
VPS_IP="your-vps-ip"
DEPLOY_DIR="/var/www/skillsphere"

echo "🚀 Deploying Skillsphere to production..."

# 1. Build
echo "1️⃣ Building binary..."
make build-prod

# 2. Copy migrations
echo "2️⃣ Copying migrations..."
scp -r internal/database/migrations "$VPS_USER@$VPS_IP:$DEPLOY_DIR/"

# 3. Run migrations
echo "3️⃣ Running migrations..."
ssh "$VPS_USER@$VPS_IP" "$DEPLOY_DIR/migrate.sh"

# 4. Copy binary
echo "4️⃣ Copying binary..."
scp bin/server "$VPS_USER@$VPS_IP:$DEPLOY_DIR/server.new"

# 5. Deploy
echo "5️⃣ Deploying..."
ssh "$VPS_USER@$VPS_IP" << 'EOF'
  cd /var/www/skillsphere
  sudo systemctl stop skillsphere
  mv server server.old
  mv server.new server
  chmod +x server
  sudo systemctl start skillsphere
  sleep 5
  sudo systemctl status skillsphere
EOF

# 6. Health check
echo "6️⃣ Health check..."
sleep 10
curl -f https://yourdomain.com/health

echo "✅ Deployment complete!"

# 7. Cleanup
echo "7️⃣ Cleaning up old binary..."
ssh "$VPS_USER@$VPS_IP" "rm -f $DEPLOY_DIR/server.old"

echo "🎉 Done!"
```

Make it executable:

```bash
chmod +x scripts/deploy.sh
```

---

## 9. Monitoring & Maintenance

### 9.1 Database Backup Script

Create `/home/skillsphere/database/backup.sh`:

```bash
#!/bin/bash
set -e

# Configuration
BACKUP_DIR="/home/skillsphere/database/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/skillsphere_$TIMESTAMP.sql"
DB_PASSWORD=$(grep DB_PASSWORD /home/skillsphere/database/.env | cut -d'=' -f2)

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Create backup
echo "Creating database backup..."
docker exec skillsphere_postgres pg_dump \
  -U skillsphere \
  -d skillsphere_prod \
  > "$BACKUP_FILE"

# Compress
gzip "$BACKUP_FILE"

# Keep only last 7 backups
find "$BACKUP_DIR" -name "*.sql.gz" -mtime +7 -delete

echo "Backup complete: ${BACKUP_FILE}.gz"
```

Make it executable:

```bash
chmod +x /home/skillsphere/database/backup.sh
```

### 9.2 Automated Backups with Cron

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * /home/skillsphere/database/backup.sh >> /var/log/skillsphere-backup.log 2>&1
```

### 9.3 Log Rotation

Create `/etc/logrotate.d/skillsphere`:

```
/var/log/caddy/skillsphere.log {
    daily
    missingok
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 caddy caddy
    sharedscripts
    postrotate
        systemctl reload caddy > /dev/null 2>&1 || true
    endscript
}
```

### 9.4 Health Check Monitoring

Create `/home/skillsphere/scripts/healthcheck.sh`:

```bash
#!/bin/bash

URL="https://yourdomain.com/health"
TIMEOUT=10

# Check health endpoint
if ! curl -f -s --max-time $TIMEOUT "$URL" > /dev/null; then
    echo "Health check failed! Restarting service..."
    sudo systemctl restart skillsphere

    # Send notification (optional)
    # curl -X POST "https://api.telegram.org/bot$TELEGRAM_TOKEN/sendMessage" \
    #   -d chat_id=$CHAT_ID \
    #   -d text="Skillsphere health check failed. Service restarted."
fi
```

Add to crontab (check every 5 minutes):

```bash
*/5 * * * * /home/skillsphere/scripts/healthcheck.sh
```

---

## 10. Rollback Procedures

### 10.1 Quick Rollback

If deployment fails:

```bash
# SSH into VPS
ssh skillsphere@your-vps-ip

# Stop service
sudo systemctl stop skillsphere

# Restore old binary
cd /var/www/skillsphere
mv server server.failed
mv server.old server

# Start service
sudo systemctl start skillsphere

# Verify
sudo systemctl status skillsphere
curl -I https://yourdomain.com/health
```

### 10.2 Database Rollback

```bash
# Stop application
sudo systemctl stop skillsphere

# Rollback last migration
export DATABASE_URL="postgres://skillsphere:your-db-password@127.0.0.1:5432/skillsphere_prod?sslmode=disable"
goose -dir /var/www/skillsphere/migrations postgres "$DATABASE_URL" down

# Restore from backup (if needed)
cd /home/skillsphere/database/backups
gunzip -k skillsphere_20260124_020000.sql.gz
docker exec -i skillsphere_postgres psql -U skillsphere -d skillsphere_prod < skillsphere_20260124_020000.sql

# Start application
sudo systemctl start skillsphere
```

### 10.3 Full System Restore

```bash
# 1. Restore database from backup
cd /home/skillsphere/database/backups
gunzip -k latest_backup.sql.gz
docker exec -i skillsphere_postgres psql -U skillsphere -d skillsphere_prod < latest_backup.sql

# 2. Restore application binary
# (Keep previous versions in /var/www/skillsphere/releases/)

# 3. Restart services
sudo systemctl restart skillsphere
sudo systemctl restart caddy
```

---

## Quick Reference

### Essential Commands

```bash
# View app logs
sudo journalctl -u skillsphere -f

# View Caddy logs
sudo journalctl -u caddy -f

# Restart app
sudo systemctl restart skillsphere

# Restart Caddy
sudo systemctl restart caddy

# Check database
docker exec -it skillsphere_postgres psql -U skillsphere -d skillsphere_prod

# Run migrations
/var/www/skillsphere/migrate.sh

# Backup database
/home/skillsphere/database/backup.sh

# Deploy new version
./scripts/deploy.sh
```

### Troubleshooting

```bash
# App won't start
sudo journalctl -u skillsphere -n 100 --no-pager
sudo systemctl status skillsphere

# Database connection issues
docker-compose -f /home/skillsphere/database/docker-compose.yml ps
docker logs skillsphere_postgres

# SSL certificate issues
sudo journalctl -u caddy -n 100 --no-pager
sudo caddy validate --config /etc/caddy/Caddyfile

# HTTP/3 not working
# Check UDP port 443
sudo ufw status
sudo netstat -ulnp | grep 443
```

---

## Security Checklist

- [ ] SSH key authentication (no password login)
- [ ] Firewall configured (UFW)
- [ ] fail2ban installed and configured
- [ ] Database only accessible from localhost
- [ ] Strong database password
- [ ] Strong admin password
- [ ] HTTPS enforced (HTTP → HTTPS redirect)
- [ ] HTTP/3 enabled
- [ ] Security headers configured
- [ ] Regular backups scheduled
- [ ] Log rotation configured
- [ ] Non-root user for application
- [ ] Systemd security hardening enabled
- [ ] Regular system updates scheduled

---

## Cost Estimation

**Hetzner VPS (CPX21):**
- 2 vCPU, 4GB RAM, 80GB SSD
- €4.15/month (~$4.50/month)

**Digital Ocean (Basic Droplet):**
- 2 vCPU, 4GB RAM, 80GB SSD
- $12/month

**Domain:**
- $10-15/year

**Total monthly cost:**
- Hetzner: ~$5/month + $1.25/month domain = **$6.25/month**
- Digital Ocean: ~$12/month + $1.25/month domain = **$13.25/month**

---

**Next Steps:**

1. Set up VPS
2. Configure domain DNS
3. Deploy database
4. Deploy application
5. Configure monitoring
6. Set up backups

Good luck with your production deployment! 🚀