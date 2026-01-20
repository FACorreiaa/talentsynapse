# Skillsphere PWA - Deployment Guide

This guide covers two production deployment approaches for your PWA.

> **⚠️ IMPORTANT: Domain Configuration**
>
> Throughout this guide, you'll see `your-domain.com` as a placeholder.
> **You MUST replace this with your actual domain name** in all configuration files and commands.
>
> For example, if your domain is `myapp.com`, replace:
> - `your-domain.com` → `myapp.com`
> - `admin@your-domain.com` → `admin@myapp.com`
> - `www.your-domain.com` → `www.myapp.com`

---

## Prerequisites

Before deploying, ensure:

1. **DNS configured**: Point your domain to server IP
2. **Ports open**: 80 (HTTP), 443 (HTTPS), 5432 (PostgreSQL - internal only)
3. **Database ready**: PostgreSQL instance accessible
4. **Binary built**: Run `make build-prod` or `./scripts/build-prod.sh`

---

## Path A: Pure Go (Built-in TLS) ⚡

**When to use**: Minimalist deployment, single binary, no reverse proxy.

### Pros:
- ✅ Single binary deployment
- ✅ No external dependencies
- ✅ Automatic Let's Encrypt certificates
- ✅ True "ship the binary" philosophy

### Cons:
- ⚠️ Requires root or `CAP_NET_BIND_SERVICE` (ports 80/443)
- ⚠️ 5-10 second downtime on binary restart
- ⚠️ Manual cache-control header management in Go code
- ⚠️ First request slow (~2-5 seconds) while obtaining cert

---

### Deployment Steps (Path A)

#### 1. Build the binary

```bash
./scripts/build-prod.sh
```

#### 2. Create deployment directory on server

```bash
ssh user@your-server.com
sudo mkdir -p /var/www/myapp
sudo mkdir -p /var/www/.cache  # For Let's Encrypt certificates
sudo chown -R $USER:$USER /var/www
```

#### 3. Copy binary to server

```bash
scp bin/server user@your-server.com:/var/www/myapp/
```

#### 4. Create systemd service

> **📝 Note:** The service file is named `myapp.service` as an example.
> You can rename it to match your project (e.g., `skillsphere.service`, `myproject.service`).
> Just remember to use the same name in all systemctl commands below.

On your server, create `/etc/systemd/system/myapp.service`:

```ini
[Unit]
Description=Skillsphere PWA (rename to your app)
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/var/www/myapp

# Environment variables
Environment="GO_ENV=production"
Environment="USE_TLS=true"
Environment="DOMAIN=your-domain.com"
Environment="DATABASE_URL=postgres://user:pass@localhost:5432/myapp?sslmode=disable"
Environment="ADMIN_EMAIL=admin@your-domain.com"

# The binary
ExecStart=/var/www/myapp/server

# Restart policy
Restart=always
RestartSec=10

# Security hardening (optional)
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/var/www/.cache

# Give capability to bind to ports 80/443 without root
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
```

#### 5. Enable and start service

```bash
sudo systemctl daemon-reload
sudo systemctl enable myapp
sudo systemctl start myapp
sudo systemctl status myapp
```

#### 6. Check logs

```bash
sudo journalctl -u myapp -f
```

**Expected output:**
```
🔒 Starting with built-in TLS for domain: your-domain.com
⚠️  Note: Binary must run as root or have CAP_NET_BIND_SERVICE
⚠️  Ensure port 80 and 443 are open and DNS is configured
🔓 HTTP server starting on :80 (ACME challenge + redirect)
🔒 HTTPS server starting on :443 (domain: your-domain.com)
```

#### 7. Test deployment

```bash
curl -I https://your-domain.com/health
# Should return: 200 OK with valid SSL certificate
```

---

## Path B: Caddy + Go (Recommended) 🎯

**When to use**: Production deployments, need zero-downtime updates, want better caching.

### Pros:
- ✅ Zero-downtime deployments
- ✅ Automatic Let's Encrypt certificates
- ✅ Built-in HTTP/2, HTTP/3 (QUIC)
- ✅ Optimized static asset serving
- ✅ Better PWA cache-control headers
- ✅ No root privileges needed for Go binary
- ✅ Custom error pages during maintenance

### Cons:
- ⚠️ Two binaries to manage (Caddy + Go)
- ⚠️ Slightly more complex setup (but minimal)

---

### Deployment Steps (Path B)

#### 1. Build the binary

```bash
./scripts/build-prod.sh
```

#### 2. Install Caddy on server

```bash
ssh user@your-server.com

# Install Caddy (Debian/Ubuntu)
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy
```

#### 3. Copy files to server

```bash
scp bin/server user@your-server.com:/var/www/myapp/
scp Caddyfile user@your-server.com:/etc/caddy/Caddyfile
```

#### 4. Create systemd service for Go app

> **📝 Note:** The service file is named `myapp.service` as an example.
> You can rename it to match your project (e.g., `skillsphere.service`, `myproject.service`).
> Just remember to use the same name in all systemctl commands below.

Create `/etc/systemd/system/myapp.service`:

```ini
[Unit]
Description=Skillsphere PWA (rename to your app) Backend
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/var/www/myapp

# Environment variables
Environment="GO_ENV=production"
Environment="PORT=8080"
Environment="DATABASE_URL=postgres://user:pass@localhost:5432/myapp?sslmode=disable"

# The binary (no TLS - Caddy handles it)
ExecStart=/var/www/myapp/server

# Restart policy
Restart=always
RestartSec=5

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict

[Install]
WantedBy=multi-user.target
```

#### 5. Update Caddyfile

Edit `/etc/caddy/Caddyfile` and replace `your-domain.com` with your actual domain:

```caddyfile
{
    email admin@your-domain.com  # Change this!
}

your-domain.com {
    encode gzip zstd

    header /assets/* {
        Cache-Control "public, max-age=31536000, immutable"
    }

    header /assets/static/sw.js {
        Cache-Control "no-cache, no-store, must-revalidate"
    }

    header /assets/static/manifest.json {
        Cache-Control "public, max-age=86400"
    }

    reverse_proxy localhost:8080 {
        fail_duration 30s
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-Proto {scheme}
    }
}

www.your-domain.com {
    redir https://your-domain.com{uri} permanent
}
```

#### 6. Enable and start services

```bash
# Start Go app
sudo systemctl daemon-reload
sudo systemctl enable myapp
sudo systemctl start myapp

# Restart Caddy with new config
sudo systemctl restart caddy

# Check status
sudo systemctl status myapp
sudo systemctl status caddy
```

#### 7. Check logs

```bash
# Go app logs
sudo journalctl -u myapp -f

# Caddy logs
sudo journalctl -u caddy -f
```

#### 8. Test deployment

```bash
# Test SSL
curl -I https://your-domain.com/health

# Test Service Worker
curl -I https://your-domain.com/assets/static/sw.js
# Should have: Cache-Control: no-cache, no-store, must-revalidate

# Test static assets
curl -I https://your-domain.com/assets/css/output.css
# Should have: Cache-Control: public, max-age=31536000, immutable
```

---

## Zero-Downtime Deployment (Path B Only)

When you want to deploy a new version:

```bash
# 1. Build new binary locally
./scripts/build-prod.sh

# 2. Copy to server (different name first)
scp bin/server user@your-server.com:/var/www/myapp/server.new

# 3. On server, atomic swap
ssh user@your-server.com
cd /var/www/myapp
mv server server.old
mv server.new server
sudo systemctl restart myapp

# 4. Check it started successfully
sudo systemctl status myapp

# 5. If OK, delete old binary
rm server.old
```

**During the restart (5-10 seconds), Caddy will:**
- Buffer incoming requests briefly
- Retry failed backend connections
- Prevent users from seeing connection errors

---

## Environment Variables Reference

### Required for all deployments:
```bash
GO_ENV=production                     # Enable production mode
DATABASE_URL=postgres://...           # PostgreSQL connection string
```

### Path A specific (Pure Go):
```bash
USE_TLS=true                          # Enable built-in Let's Encrypt
DOMAIN=your-domain.com                # Your domain name
ADMIN_EMAIL=admin@your-domain.com     # For Let's Encrypt notifications
```

### Path B specific (Caddy):
```bash
PORT=8080                             # Backend port (Caddy proxies to this)
```

---

## Firewall Configuration

### Path A (Pure Go):
```bash
sudo ufw allow 80/tcp    # HTTP (ACME challenge)
sudo ufw allow 443/tcp   # HTTPS
sudo ufw allow 5432/tcp  # PostgreSQL (internal only - restrict to localhost)
sudo ufw enable
```

### Path B (Caddy):
```bash
sudo ufw allow 80/tcp    # HTTP (handled by Caddy)
sudo ufw allow 443/tcp   # HTTPS (handled by Caddy)
# Port 8080 stays internal (no firewall rule needed)
sudo ufw enable
```

---

## Monitoring & Maintenance

### Health Check Endpoint

Both paths expose `/health`:

```bash
curl https://your-domain.com/health
```

**Expected response:**
```json
{
  "status": "healthy",
  "total_conns": "25",
  "acquired_conns": "5",
  "idle_conns": "20",
  "constructing": "0",
  "max_conns": "25"
}
```

### Set up monitoring

Use a service like UptimeRobot, Pingdom, or self-hosted solution:

```bash
# Check every 5 minutes
* */5 * * * curl -f https://your-domain.com/health || systemctl restart myapp
```

### View logs

```bash
# Path A
sudo journalctl -u myapp -f --lines=100

# Path B
sudo journalctl -u myapp -f --lines=100
sudo journalctl -u caddy -f --lines=100
```

---

## Backup & Restore

### Database backup

```bash
# Backup
pg_dump -U myapp -h localhost myapp > backup-$(date +%Y%m%d).sql

# Restore
psql -U myapp -h localhost myapp < backup-20260119.sql
```

### Binary backup

```bash
# Before deploying new version
cp /var/www/myapp/server /var/www/myapp/server.backup-$(date +%Y%m%d)
```

---

## Troubleshooting

### Issue: "Certificate obtain failed"

**Path A Only**
```bash
# Check DNS is pointing to server
dig your-domain.com

# Check ports are accessible
sudo netstat -tlnp | grep -E ':(80|443)'

# Check Let's Encrypt rate limits
# You get 5 failed attempts per hour per domain
```

### Issue: "Connection refused"

```bash
# Check service is running
sudo systemctl status myapp

# Check it's listening on correct port
sudo netstat -tlnp | grep server

# Check database connection
psql -U myapp -h localhost myapp -c "SELECT 1;"
```

### Issue: "Service Worker won't register"

```bash
# MUST be HTTPS (except localhost)
# Check SSL certificate is valid
curl -vI https://your-domain.com 2>&1 | grep -E "SSL|TLS"

# Check sw.js has correct headers
curl -I https://your-domain.com/assets/static/sw.js
# Must have: Cache-Control: no-cache, no-store, must-revalidate
```

---

## Performance Tuning

### Database Connection Pool

Edit `internal/database/database.go`:

```go
// For small VPS (1-2 CPU cores)
config.MaxConns = 10
config.MinConns = 2

// For medium server (4 CPU cores)
config.MaxConns = 25
config.MinConns = 5

// For large server (8+ CPU cores)
config.MaxConns = 50
config.MinConns = 10
```

### OS-level tuning

```bash
# Increase file descriptor limit
sudo vim /etc/security/limits.conf
# Add:
* soft nofile 65535
* hard nofile 65535

# TCP tuning for high concurrency
sudo vim /etc/sysctl.conf
# Add:
net.core.somaxconn = 1024
net.ipv4.tcp_max_syn_backlog = 2048
net.ipv4.ip_local_port_range = 10000 65535

# Apply
sudo sysctl -p
```

---

## Security Checklist

- [ ] Firewall configured (ufw or iptables)
- [ ] Database not exposed to internet
- [ ] Strong database password
- [ ] Regular security updates (`sudo apt update && sudo apt upgrade`)
- [ ] Fail2ban installed for SSH protection
- [ ] Non-root user for application (Path B)
- [ ] Regular database backups
- [ ] Log rotation configured
- [ ] HTTPS enforced (no HTTP access)
- [ ] Security headers enabled (already in `routes.go`)

---

## My Recommendation for Skillsphere

**Use Path B (Caddy + Go)** because:

1. ✅ Better for PWA deployments (proper cache headers)
2. ✅ Zero-downtime updates
3. ✅ Minimal complexity (one Caddyfile)
4. ✅ Still close to "single binary" philosophy (Caddy is also Go)
5. ✅ HTTP/3 support out of the box
6. ✅ Better production experience overall

**Use Path A (Pure Go)** only if:
- You're deploying to very constrained environment
- You absolutely need single-file deployment
- You're comfortable with brief downtime on updates

---

## Quick Start (TL;DR)

### Path B (Recommended):

```bash
# 1. Build
./scripts/build-prod.sh

# 2. Deploy
scp bin/server user@server:/var/www/myapp/
scp Caddyfile user@server:/etc/caddy/Caddyfile

# 3. Configure systemd service (see above)

# 4. Start
sudo systemctl start myapp caddy

# 5. Done!
curl https://your-domain.com/health
```

---

**Need help?** Check logs first:
```bash
sudo journalctl -u myapp -f
sudo journalctl -u caddy -f
```
