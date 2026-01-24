# Security & Observability Guide

Comprehensive guide for logging, monitoring, error tracking, and security for Skillsphere PWA.

---

## Table of Contents

1. [Logging Strategy](#1-logging-strategy)
2. [Error Tracking](#2-error-tracking)
3. [Security Monitoring](#3-security-monitoring)
4. [Observability Stack](#4-observability-stack)
5. [Recommended Setup by Budget](#5-recommended-setup-by-budget)
6. [Implementation Guide](#6-implementation-guide)

---

## 1. Logging Strategy

### What to Log

#### ✅ DO Log:
- **Authentication events** (login, logout, failed attempts)
- **Authorization failures** (access denied, insufficient permissions)
- **API errors** (500s, database errors, external API failures)
- **Business events** (user registration, skill matches, payments)
- **Security events** (rate limit exceeded, suspicious activity)
- **Performance issues** (slow queries, high memory usage)

#### ❌ DON'T Log:
- Passwords (even hashed)
- Session tokens
- API keys
- Credit card numbers
- Personal identifiable information (PII) without hashing
- Full request bodies (may contain secrets)

### Log Levels

```go
// Use structured logging with levels
log.Debug("Cache hit", "key", key)                      // Development only
log.Info("User logged in", "user_id", userID)           // Normal operations
log.Warn("Rate limit approaching", "ip", ip)            // Potential issues
log.Error("Database query failed", "error", err)        // Errors
log.Fatal("Cannot connect to database", "error", err)   // Critical failures
```

---

## 2. Error Tracking & Performance Profiling

### ✅ Implemented: Sentry

Sentry is **already configured** in this project. See `docs/SENTRY_USAGE_GUIDE.md` for usage examples.

**What's configured:**
- ✅ Automatic error capture from panics
- ✅ HTTP request context (method, URL, headers)
- ✅ Performance tracing (100% sample rate)
- ✅ Environment-aware (development/production)
- ✅ Helper functions in `internal/server/sentry_helpers.go`

**Location:** `cmd/server/main.go:16-27`, `internal/server/routes.go:32-39`

### ✅ Implemented: pprof Profiling

Go's built-in profiler is **already configured** with production-safe authentication. See `docs/PPROF_PROFILING_GUIDE.md` for detailed usage.

**What's configured:**
- ✅ CPU profiling
- ✅ Memory (heap) profiling
- ✅ Goroutine profiling
- ✅ Mutex and block profiling
- ✅ Production authentication (Basic Auth)
- ✅ Environment-aware security

**Endpoints:** Available at `/debug/pprof/*`
**Location:** `internal/server/routes.go:192-208`, `internal/server/pprof_middleware.go`

**Quick usage:**
```bash
# Development (no auth required)
go tool pprof http://localhost:8081/debug/pprof/profile?seconds=30

# Production (requires PPROF_USERNAME and PPROF_PASSWORD)
go tool pprof http://admin:password@your-server.com/debug/pprof/profile?seconds=30
```

### Comparison: Sentry vs Alternatives

| Feature | Sentry | Rollbar | Bugsnag | New Relic Errors |
|---------|--------|---------|---------|------------------|
| **Price (Free Tier)** | 5K events/month | 5K events/month | 7.5K events/month | Included with APM |
| **Go Support** | ✅ Excellent | ✅ Good | ✅ Good | ✅ Excellent |
| **Source Maps** | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes |
| **Performance Tracing** | ✅ Yes | ❌ No | ❌ No | ✅ Yes |
| **Alerting** | ✅ Good | ✅ Good | ✅ Good | ✅ Excellent |
| **Cost (Paid)** | $26/month | $25/month | $59/month | $99/month |

### **Recommendation: Sentry**

**Why Sentry:**
- ✅ Best Go SDK
- ✅ Excellent error grouping
- ✅ Performance monitoring included
- ✅ Generous free tier
- ✅ Great user context tracking
- ✅ Release tracking

**When to use alternatives:**
- **Rollbar**: If you need cheaper paid tier
- **Bugsnag**: If you use Atlassian products
- **New Relic**: If you already use New Relic APM

---

## 3. Security Monitoring

### What to Monitor

#### Authentication Security
```go
// Failed login attempts
if loginFailed {
    log.Warn("Failed login attempt",
        "email", email,
        "ip", r.RemoteAddr,
        "user_agent", r.UserAgent())

    // Track in metrics
    metrics.Inc("auth_failed_attempts", map[string]string{
        "ip": r.RemoteAddr,
    })
}

// Account lockout
if accountLocked {
    log.Error("Account locked due to failed attempts",
        "user_id", userID,
        "ip", r.RemoteAddr)

    // Alert security team
    alert.Send("Account Lockout", userID)
}
```

#### Suspicious Activity
```go
// Multiple IPs for same user
if suspiciousIPChange(userID, newIP) {
    log.Warn("Suspicious IP change detected",
        "user_id", userID,
        "old_ip", oldIP,
        "new_ip", newIP)
}

// Rate limiting
if rateLimitExceeded {
    log.Warn("Rate limit exceeded",
        "ip", r.RemoteAddr,
        "endpoint", r.URL.Path,
        "limit", limit)
}

// SQL injection attempts
if containsSQLInjection(input) {
    log.Error("Possible SQL injection attempt",
        "ip", r.RemoteAddr,
        "input", sanitize(input))
}
```

#### Data Access
```go
// Admin actions
if isAdminAction {
    log.Info("Admin action performed",
        "admin_id", adminID,
        "action", action,
        "target_user_id", targetUserID)
}

// Bulk data access
if bulkDataAccess {
    log.Warn("Bulk data access",
        "user_id", userID,
        "records_count", count,
        "endpoint", endpoint)
}
```

---

## 4. Observability Stack

### Comparison: Full Observability Platforms

| Solution | Best For | Cost | Complexity | Self-Hosted |
|----------|----------|------|------------|-------------|
| **Sentry** | Error tracking | Free - $26/mo | Low | No |
| **New Relic** | Full observability | $99/mo | Medium | No |
| **Datadog** | Enterprise | $15/host/mo | High | No |
| **Prometheus + Grafana** | Self-hosted | Free | High | Yes |
| **Loki + Grafana** | Logs | Free | Medium | Yes |
| **ELK Stack** | Logs | Free | Very High | Yes |
| **Grafana Cloud** | Hybrid | Free - $49/mo | Medium | Partial |

### **Recommendation by Use Case**

#### **Startup / MVP (You are here)**
```
✅ Sentry (Errors + Performance)
✅ Structured Logs to file (Loki later)
✅ Basic metrics in-app
✅ Caddy access logs

Cost: Free - $26/month
```

#### **Growing (1K+ users)**
```
✅ Sentry (Errors)
✅ Grafana Cloud (Metrics + Logs)
✅ Uptime monitoring (UptimeRobot)

Cost: ~$49/month
```

#### **Scale (10K+ users)**
```
✅ New Relic or Datadog (Full observability)
✅ PagerDuty (Alerting)
✅ Security monitoring (Snyk, OWASP)

Cost: $200-500/month
```

---

## 5. Recommended Setup by Budget

### **Free Tier Setup** (Recommended for you)

**What you get:**
- Error tracking (Sentry: 5K events/month)
- Logs (Self-hosted: systemd journald)
- Metrics (In-app: /health endpoint)
- Uptime (UptimeRobot: 50 monitors free)

**Stack:**
```
Errors:    Sentry (Free)
Logs:      systemd journald → Loki (optional)
Metrics:   /health endpoint → Prometheus (optional)
Uptime:    UptimeRobot (Free)
Alerts:    Email/Slack webhooks
```

**Monthly cost: $0**

### **Paid Setup ($50/month budget)**

```
Errors:    Sentry Team ($26/mo)
Logs:      Grafana Cloud Logs ($8/mo)
Metrics:   Grafana Cloud Metrics ($8/mo)
Uptime:    UptimeRobot Pro ($7/mo)
Total:     ~$49/month
```

### **Production Setup ($200/month budget)**

```
Full Stack: New Relic Pro ($99/mo)
Errors:     Sentry Team ($26/mo)
Security:   Snyk ($25/mo)
Uptime:     Better Uptime ($20/mo)
Alerts:     PagerDuty ($21/mo)
Total:      ~$191/month
```

---

## 6. Implementation Guide

### Phase 1: Free Tier (Implement Now)

#### 1. Structured Logging

Create `internal/logger/logger.go`:

```go
package logger

import (
    "context"
    "log/slog"
    "os"
)

var Logger *slog.Logger

func Init(env string) {
    var handler slog.Handler

    if env == "production" {
        // JSON logs for production (easy to parse)
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelInfo,
            AddSource: true,
        })
    } else {
        // Pretty logs for development
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelDebug,
        })
    }

    Logger = slog.New(handler)
    slog.SetDefault(Logger)
}

// Helper functions
func Info(msg string, args ...any) {
    Logger.Info(msg, args...)
}

func Error(msg string, args ...any) {
    Logger.Error(msg, args...)
}

func Warn(msg string, args ...any) {
    Logger.Warn(msg, args...)
}

func Debug(msg string, args ...any) {
    Logger.Debug(msg, args...)
}

// WithContext adds request context to logs
func WithContext(ctx context.Context) *slog.Logger {
    requestID := ctx.Value("request_id")
    userID := ctx.Value("user_id")

    return Logger.With(
        "request_id", requestID,
        "user_id", userID,
    )
}
```

#### 2. Sentry Integration

Install Sentry SDK:

```bash
go get github.com/getsentry/sentry-go
```

Create `internal/monitoring/sentry.go`:

```go
package monitoring

import (
    "fmt"
    "log"
    "os"
    "time"

    "github.com/getsentry/sentry-go"
)

func InitSentry() error {
    dsn := os.Getenv("SENTRY_DSN")
    env := os.Getenv("GO_ENV")

    if dsn == "" {
        log.Println("⚠️  SENTRY_DSN not set, skipping Sentry initialization")
        return nil
    }

    err := sentry.Init(sentry.ClientOptions{
        Dsn:              dsn,
        Environment:      env,
        Release:          fmt.Sprintf("skillsphere@%s", os.Getenv("VERSION")),
        TracesSampleRate: 0.2, // 20% of transactions for performance monitoring
        Debug:            env == "development",
        BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
            // Don't send events in development
            if env == "development" {
                return nil
            }
            return event
        },
    })

    if err != nil {
        return fmt.Errorf("sentry initialization failed: %w", err)
    }

    log.Println("✅ Sentry initialized")
    return nil
}

// CaptureError sends error to Sentry with context
func CaptureError(err error, ctx map[string]interface{}) {
    sentry.WithScope(func(scope *sentry.Scope) {
        scope.SetContext("custom", ctx)
        sentry.CaptureException(err)
    })
}

// CaptureMessage sends message to Sentry
func CaptureMessage(message string, level sentry.Level) {
    sentry.CaptureMessage(message)
}

// Flush waits for events to be sent (use on shutdown)
func Flush() {
    sentry.Flush(2 * time.Second)
}
```

#### 3. Security Middleware

Create `internal/middleware/security.go`:

```go
package middleware

import (
    "log/slog"
    "net/http"
    "strings"
    "time"

    "skillsphere/internal/logger"
)

// SecurityLogger logs security-relevant events
func SecurityLogger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Detect suspicious patterns
        if isSuspiciousRequest(r) {
            logger.Warn("Suspicious request detected",
                "ip", r.RemoteAddr,
                "path", r.URL.Path,
                "user_agent", r.UserAgent(),
            )
        }

        // Wrap response writer to capture status code
        wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

        next.ServeHTTP(wrapped, r)

        // Log failed auth attempts
        if r.URL.Path == "/auth/login" && wrapped.statusCode == 401 {
            logger.Warn("Failed login attempt",
                "ip", r.RemoteAddr,
                "user_agent", r.UserAgent(),
            )
        }

        // Log slow requests
        duration := time.Since(start)
        if duration > 1*time.Second {
            logger.Warn("Slow request",
                "path", r.URL.Path,
                "duration_ms", duration.Milliseconds(),
            )
        }

        // Log all requests in JSON format
        logger.Info("HTTP request",
            "method", r.Method,
            "path", r.URL.Path,
            "status", wrapped.statusCode,
            "duration_ms", duration.Milliseconds(),
            "ip", r.RemoteAddr,
        )
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
    w.statusCode = code
    w.ResponseWriter.WriteHeader(code)
}

func isSuspiciousRequest(r *http.Request) bool {
    // SQL injection patterns
    sqlPatterns := []string{
        "' OR '1'='1",
        "'; DROP TABLE",
        "UNION SELECT",
        "<script>",
        "javascript:",
    }

    query := r.URL.Query().Encode()
    for _, pattern := range sqlPatterns {
        if strings.Contains(strings.ToUpper(query), strings.ToUpper(pattern)) {
            return true
        }
    }

    // Path traversal
    if strings.Contains(r.URL.Path, "../") {
        return true
    }

    return false
}
```

#### 4. Health Check with Metrics

Update your health endpoint:

```go
// internal/handlers/health.go
package handlers

import (
    "encoding/json"
    "net/http"
    "runtime"
    "time"

    "skillsphere/internal/database"
)

type HealthResponse struct {
    Status      string            `json:"status"`
    Version     string            `json:"version"`
    Timestamp   time.Time         `json:"timestamp"`
    Database    DatabaseHealth    `json:"database"`
    System      SystemHealth      `json:"system"`
}

type DatabaseHealth struct {
    Status         string `json:"status"`
    TotalConns     int32  `json:"total_conns"`
    AcquiredConns  int32  `json:"acquired_conns"`
    IdleConns      int32  `json:"idle_conns"`
    MaxConns       int32  `json:"max_conns"`
}

type SystemHealth struct {
    Goroutines    int    `json:"goroutines"`
    MemoryMB      uint64 `json:"memory_mb"`
    Uptime        string `json:"uptime"`
}

var startTime = time.Now()

func Health(db *database.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Check database
        stats := db.Pool.Stat()
        dbStatus := "healthy"

        if stats.AcquiredConns() >= stats.MaxConns() {
            dbStatus = "degraded"
        }

        // System metrics
        var m runtime.MemStats
        runtime.ReadMemStats(&m)

        health := HealthResponse{
            Status:    "healthy",
            Version:   os.Getenv("VERSION"),
            Timestamp: time.Now(),
            Database: DatabaseHealth{
                Status:        dbStatus,
                TotalConns:    stats.TotalConns(),
                AcquiredConns: stats.AcquiredConns(),
                IdleConns:     stats.IdleConns(),
                MaxConns:      stats.MaxConns(),
            },
            System: SystemHealth{
                Goroutines: runtime.NumGoroutine(),
                MemoryMB:   m.Alloc / 1024 / 1024,
                Uptime:     time.Since(startTime).String(),
            },
        }

        if dbStatus == "degraded" {
            health.Status = "degraded"
            w.WriteHeader(http.StatusServiceUnavailable)
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(health)
    }
}
```

#### 5. Update main.go

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "skillsphere/internal/logger"
    "skillsphere/internal/monitoring"
    "skillsphere/internal/database"
    "skillsphere/internal/server"
)

func main() {
    // Initialize logger
    logger.Init(os.Getenv("GO_ENV"))

    // Initialize Sentry
    if err := monitoring.InitSentry(); err != nil {
        logger.Error("Failed to initialize Sentry", "error", err)
    }
    defer monitoring.Flush()

    // Connect to database
    ctx := context.Background()
    db, err := database.New(ctx, os.Getenv("DATABASE_URL"))
    if err != nil {
        logger.Error("Failed to connect to database", "error", err)
        monitoring.CaptureError(err, map[string]interface{}{
            "component": "database",
        })
        os.Exit(1)
    }
    defer db.Close()

    // Start server
    srv := server.New(db)

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-quit
        logger.Info("Shutting down server...")

        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        if err := srv.Shutdown(ctx); err != nil {
            logger.Error("Server forced to shutdown", "error", err)
        }
    }()

    logger.Info("Server starting",
        "port", os.Getenv("PORT"),
        "env", os.Getenv("GO_ENV"),
    )

    if err := srv.Start(); err != nil {
        logger.Error("Server failed to start", "error", err)
        monitoring.CaptureError(err, map[string]interface{}{
            "component": "server",
        })
        os.Exit(1)
    }
}
```

#### 6. Environment Configuration

Add to `.env`:

```bash
# Monitoring
SENTRY_DSN=https://your-sentry-dsn@sentry.io/project-id
VERSION=1.0.0

# Optional: Slack webhook for alerts
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

#### 7. UptimeRobot Setup

1. Go to https://uptimerobot.com
2. Create free account
3. Add HTTP(s) monitor:
   - URL: `https://yourdomain.com/health`
   - Interval: 5 minutes
   - Alert when: Status is not 200 or response contains "degraded"
4. Add notification channels (Email, Slack, etc.)

---

### Phase 2: When You Have Budget

#### Option A: Grafana Cloud (Recommended)

**Cost: ~$49/month**

```bash
# Install Grafana Agent
wget https://github.com/grafana/agent/releases/latest/download/grafana-agent-linux-amd64
sudo mv grafana-agent-linux-amd64 /usr/local/bin/grafana-agent
sudo chmod +x /usr/local/bin/grafana-agent

# Configure agent (prometheus metrics + loki logs)
sudo vim /etc/grafana-agent.yaml
```

#### Option B: Self-Hosted Prometheus + Loki

**Cost: $0 (but more work)**

See separate guide: `SELF_HOSTED_MONITORING.md`

---

## Security Best Practices

### 1. Rate Limiting

```go
import "github.com/didip/tollbooth/v7"

// Rate limit: 20 requests per minute per IP
limiter := tollbooth.NewLimiter(20, nil)
router.Use(tollbooth.LimitHandler(limiter, next))
```

### 2. Request ID Tracking

```go
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := uuid.New().String()
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        w.Header().Set("X-Request-ID", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 3. Audit Logging

```go
// Log all admin actions
func AuditLog(adminID, action, targetID string) {
    logger.Info("Admin action",
        "admin_id", adminID,
        "action", action,
        "target_id", targetID,
        "timestamp", time.Now(),
    )
}
```

---

## Summary

### **Currently Implemented (Free)**:
1. ✅ **Sentry** - Error tracking and performance monitoring (configured)
2. ✅ **pprof** - CPU, memory, and goroutine profiling (configured)
3. ✅ /health endpoint - Basic metrics
4. ✅ Security middleware - Rate limiting, CORS, secure headers

### **To Add (Free)**:
1. Structured logging (slog) - See examples above
2. UptimeRobot - Uptime monitoring

### **Add later ($50/mo)**:
1. Grafana Cloud (metrics + logs)
2. Sentry paid tier (more events)
3. Better Uptime (better alerting)

### **Scale further ($200/mo)**:
1. New Relic or Datadog (full observability)
2. PagerDuty (on-call management)
3. Snyk (security scanning)

---

## Quick Start

```bash
# 1. Install Sentry SDK
go get github.com/getsentry/sentry-go

# 2. Get Sentry DSN
# Visit: https://sentry.io/signup/
# Create project → Copy DSN

# 3. Add to .env
echo "SENTRY_DSN=your-dsn-here" >> .env

# 4. Update code (see examples above)

# 5. Test
make dev
# Trigger an error and check Sentry dashboard
```

---

**Next steps:**
1. Implement structured logging
2. Add Sentry integration
3. Set up UptimeRobot
4. Add security middleware
5. Monitor for a week
6. Decide on paid tier if needed
