# Deployment Configuration Checklist

Before deploying your application, you need to replace the following placeholders with your actual values:

## 🔧 Required Replacements

### 1. Domain Configuration

Replace these in **ALL** files (`Caddyfile`, `DEPLOYMENT.md`, systemd service files):

| Placeholder | Replace With | Example |
|-------------|-------------|---------|
| `your-domain.com` | Your actual domain | `myapp.com` |
| `www.your-domain.com` | Your www subdomain | `www.myapp.com` |
| `admin@your-domain.com` | Your admin email | `admin@myapp.com` |

### 2. Application Name

Replace these in systemd service files and paths:

| Placeholder | Replace With | Example |
|-------------|-------------|---------|
| `myapp` | Your application name | `skillsphere` |
| `myapp.service` | Your service file name | `skillsphere.service` |
| `/var/www/myapp/` | Your deployment path | `/var/www/skillsphere/` |

### 3. Database Configuration

Replace these in environment variables and backup scripts:

| Placeholder | Replace With | Example |
|-------------|-------------|---------|
| `postgres://user:pass@localhost:5432/myapp` | Your database connection string | `postgres://admin:secret123@localhost:5432/skillsphere` |
| `-U myapp` | Your database username | `-U skillsphere` |
| `localhost myapp` | Your database name | `localhost skillsphere` |

### 4. Server Configuration

Replace these when deploying:

| Placeholder | Replace With | Example |
|-------------|-------------|---------|
| `user@your-server.com` | Your SSH user and server | `deploy@192.168.1.100` |

---

## 📝 Files to Update

### 1. Caddyfile
```bash
# Edit these lines:
email admin@your-domain.com         → email admin@myactual.com
your-domain.com {                   → myactual.com {
www.your-domain.com {              → www.myactual.com {
```

### 2. Systemd Service Files

**For Path A** (`/etc/systemd/system/myapp.service`):
```ini
# Edit these lines:
Description=Skillsphere PWA (rename to your app)  → Description=MyApp PWA
WorkingDirectory=/var/www/myapp                   → WorkingDirectory=/var/www/myactual
Environment="DOMAIN=your-domain.com"              → Environment="DOMAIN=myactual.com"
Environment="DATABASE_URL=postgres://user:pass@localhost:5432/myapp?sslmode=disable"
  → Environment="DATABASE_URL=postgres://admin:secret@localhost:5432/myactual?sslmode=disable"
Environment="ADMIN_EMAIL=admin@your-domain.com"   → Environment="ADMIN_EMAIL=admin@myactual.com"
ExecStart=/var/www/myapp/server                   → ExecStart=/var/www/myactual/server
```

**For Path B** (`/etc/systemd/system/myapp.service`):
```ini
# Edit these lines:
Description=Skillsphere PWA (rename to your app) Backend  → Description=MyApp Backend
WorkingDirectory=/var/www/myapp                           → WorkingDirectory=/var/www/myactual
Environment="DATABASE_URL=postgres://user:pass@localhost:5432/myapp?sslmode=disable"
  → Environment="DATABASE_URL=postgres://admin:secret@localhost:5432/myactual?sslmode=disable"
ExecStart=/var/www/myapp/server                           → ExecStart=/var/www/myactual/server
```

### 3. Deployment Commands

Update all deployment commands:
```bash
# Before:
scp bin/server user@your-server.com:/var/www/myapp/

# After:
scp bin/server deploy@192.168.1.100:/var/www/myactual/
```

---

## ✅ Verification Checklist

Before starting deployment, verify you've updated:

- [ ] Domain name in `Caddyfile` (2 places: main domain + www redirect)
- [ ] Email address in `Caddyfile` global config
- [ ] Application name in systemd service file (`Description` field)
- [ ] Working directory in systemd service file
- [ ] Database connection string in systemd service file
- [ ] Domain name in `DOMAIN` environment variable (Path A only)
- [ ] Admin email in `ADMIN_EMAIL` environment variable (Path A only)
- [ ] Binary path in `ExecStart` directive
- [ ] Server address in deployment commands (scp, ssh)
- [ ] Service name in all `systemctl` commands

---

## 🚀 Quick Setup Script

You can use this script to help with replacements:

```bash
#!/bin/bash
# setup-deployment.sh - Configure deployment files

read -p "Enter your domain (e.g., myapp.com): " DOMAIN
read -p "Enter your app name (e.g., skillsphere): " APPNAME
read -p "Enter your admin email: " EMAIL
read -p "Enter your database user: " DBUSER
read -p "Enter your database password: " DBPASS
read -p "Enter your server SSH address (user@host): " SERVER

echo "Updating configuration files..."

# Update Caddyfile
sed -i "s/your-domain.com/$DOMAIN/g" Caddyfile
sed -i "s/admin@your-domain.com/$EMAIL/g" Caddyfile

# Create systemd service file
cat > /tmp/$APPNAME.service << EOF
[Unit]
Description=$APPNAME PWA
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/var/www/$APPNAME

Environment="GO_ENV=production"
Environment="PORT=8080"
Environment="DATABASE_URL=postgres://$DBUSER:$DBPASS@localhost:5432/$APPNAME?sslmode=disable"

ExecStart=/var/www/$APPNAME/server

Restart=always
RestartSec=5

NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict

[Install]
WantedBy=multi-user.target
EOF

echo "✅ Configuration complete!"
echo ""
echo "Next steps:"
echo "1. Review generated service file: /tmp/$APPNAME.service"
echo "2. Review Caddyfile"
echo "3. Deploy to server:"
echo "   scp bin/server $SERVER:/var/www/$APPNAME/"
echo "   scp Caddyfile $SERVER:/etc/caddy/Caddyfile"
echo "   scp /tmp/$APPNAME.service $SERVER:/tmp/"
echo "   ssh $SERVER 'sudo mv /tmp/$APPNAME.service /etc/systemd/system/'"
```

Save as `setup-deployment.sh` and run:
```bash
chmod +x setup-deployment.sh
./setup-deployment.sh
```

---

## 📚 Additional Resources

- Main deployment guide: [DEPLOYMENT.md](DEPLOYMENT.md)
- Optimization guide: [OPTIMIZATION_GUIDE.md](OPTIMIZATION_GUIDE.md)
- Caddy documentation: https://caddyserver.com/docs/
- Systemd service documentation: https://www.freedesktop.org/software/systemd/man/systemd.service.html

Perfect! I've updated all the deployment documentation to use generic placeholders instead of
skillsphere.com. Here's what changed:

📝 Files Updated:

1. Caddyfile

- ✅ skillsphere.com → your-domain.com
- ✅ Added prominent warnings to replace domain
- ✅ Updated email placeholder

2. docs/DEPLOYMENT.md

- ✅ All domain references updated to your-domain.com
- ✅ Service names changed from skillsphere.service → myapp.service
- ✅ Paths updated from /var/www/skillsphere → /var/www/myapp
- ✅ Database names updated from skillsphere → myapp
- ✅ Added clear warning box at the top with examples
- ✅ Added notes before systemd service sections

3. docs/DEPLOYMENT_CHECKLIST.md (NEW!)

Created a comprehensive checklist that includes:
- ✅ Table of all placeholders to replace
- ✅ Before/after examples
- ✅ File-by-file replacement guide
- ✅ Quick setup script to automate replacements
- ✅ Pre-deployment verification checklist

🎯 Key Placeholders You'll Need to Replace:

| Placeholder                               | What to Replace With                     |
  |-------------------------------------------|------------------------------------------|
| your-domain.com                           | Your actual domain (e.g., myapp.com)     |
| admin@your-domain.com                     | Your admin email                         |
| myapp                                     | Your app name (can keep as is or rename) |
| user@your-server.com                      | Your SSH server address                  |
| postgres://user:pass@localhost:5432/myapp | Your database credentials                |

📚 Documentation Structure:

docs/
├── DEPLOYMENT.md            # Main deployment guide (both paths)
├── DEPLOYMENT_CHECKLIST.md  # Configuration replacement guide
└── OPTIMIZATION_GUIDE.md    # Build optimization tips

🚀 Quick Start:

When you're ready to deploy:

1. Review the checklist:
   cat docs/DEPLOYMENT_CHECKLIST.md
2. Update Caddyfile:
# Replace your-domain.com with your actual domain
sed -i 's/your-domain.com/myactual.com/g' Caddyfile
sed -i 's/admin@your-domain.com/admin@myactual.com/g' Caddyfile
3. Follow deployment guide:
   cat docs/DEPLOYMENT.md

All the documentation now uses generic placeholders with clear instructions on what needs to be
changed!
