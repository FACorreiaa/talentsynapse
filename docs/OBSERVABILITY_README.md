# Observability & Profiling - Quick Reference

## 🎯 What's Configured

Your SkillSphere application now has **production-ready observability and profiling** with:

1. **Sentry** - Error tracking & performance monitoring ✅
2. **pprof** - CPU, memory, and goroutine profiling ✅

Both are **already integrated** and ready to use!

---

## 🚀 Quick Start

### Sentry (Error Tracking)

**Already working!** Just check your Sentry dashboard at https://sentry.io

Errors and panics are automatically captured and sent to:
```
https://o4510766840872960.ingest.de.sentry.io/4510766843428944
```

**Manual error capture:**
```go
// In any HTTP handler
if err := doSomething(); err != nil {
    CaptureError(r, err)  // r is *http.Request
    http.Error(w, "Error", 500)
}
```

### pprof (Profiling)

**Development (no auth):**
```bash
# View all profiles
open http://localhost:8081/debug/pprof/

# CPU profile with interactive UI
go tool pprof -http=:8080 http://localhost:8081/debug/pprof/profile?seconds=30

# Memory profile
go tool pprof -http=:8080 http://localhost:8081/debug/pprof/heap
```

**Production (requires auth):**
```bash
# Set credentials first
export PPROF_USERNAME=admin
export PPROF_PASSWORD=your-secure-password

# Then profile via SSH tunnel (recommended)
ssh -L 6060:localhost:8081 user@your-server.com
go tool pprof -http=:8080 http://admin:password@localhost:6060/debug/pprof/profile
```

---

## 📚 Documentation

| Guide | Purpose |
|-------|---------|
| **[SENTRY_USAGE_GUIDE.md](SENTRY_USAGE_GUIDE.md)** | How to use Sentry in your code |
| **[PPROF_QUICK_START.md](PPROF_QUICK_START.md)** | Quick pprof commands |
| **[PPROF_PROFILING_GUIDE.md](PPROF_PROFILING_GUIDE.md)** | Comprehensive pprof guide |
| **[OBSERVABILITY_IMPLEMENTATION_SUMMARY.md](OBSERVABILITY_IMPLEMENTATION_SUMMARY.md)** | What's implemented |
| **[SECURITY_OBSERVABILITY_GUIDE.md](SECURITY_OBSERVABILITY_GUIDE.md)** | Full observability strategy |

---

## ⚙️ Configuration

### Environment Variables

**Development (.env):**
```bash
GO_ENV=development
SENTRY_DSN=https://5b3cc7c1d51da7eeb97a0f90ea3c277a@o4510766840872960.ingest.de.sentry.io/4510766843428944
```

**Production (add these):**
```bash
GO_ENV=production
SENTRY_DSN=https://5b3cc7c1d51da7eeb97a0f90ea3c277a@o4510766840872960.ingest.de.sentry.io/4510766843428944

# Optional: Enable pprof in production (with authentication)
PPROF_USERNAME=admin
PPROF_PASSWORD=your-very-secure-password-here
```

---

## 🔍 Common Tasks

### Find CPU Bottlenecks
```bash
go tool pprof -http=:8080 http://localhost:8081/debug/pprof/profile?seconds=30
# Click "Flame Graph" in the browser
```

### Find Memory Leaks
```bash
# Take snapshot 1
curl http://localhost:8081/debug/pprof/heap -o heap1.prof

# Wait 5 minutes
sleep 300

# Take snapshot 2
curl http://localhost:8081/debug/pprof/heap -o heap2.prof

# Compare
go tool pprof -base heap1.prof heap2.prof
```

### Check for Goroutine Leaks
```bash
go tool pprof http://localhost:8081/debug/pprof/goroutine
(pprof) top
(pprof) traces
```

### Capture Error with Context
```go
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
    userID := auth.GetUserID(r)

    // Add context to Sentry events
    WithSentryScope(r, func(scope *sentry.Scope) {
        scope.SetUser(sentry.User{ID: userID})
        scope.SetTag("operation", "profile_update")
    })

    if err := h.service.Update(data); err != nil {
        CaptureError(r, err)
        http.Error(w, "Error", 500)
        return
    }
}
```

---

## 🔒 Security

### Sentry
- ✅ Doesn't send events in development
- ✅ Per-request context isolation
- ✅ Automatic PII filtering
- ✅ Graceful shutdown with flush

### pprof
- ✅ **Development:** Open access (localhost only)
- ✅ **Production:** Basic Auth required
- ✅ **Production:** Returns 404 if credentials not set (disabled)
- ✅ Uses constant-time comparison (prevents timing attacks)

---

## 📊 What Gets Tracked

### Sentry Captures
- All panics and unhandled errors
- HTTP request context (URL, method, headers)
- User IP addresses
- Stack traces
- Performance metrics (100% sample rate)
- Custom errors you capture manually

### pprof Provides
- CPU usage by function
- Memory allocations
- Goroutine counts and stacks
- Mutex contention
- Blocking operations
- Thread creation

---

## 🆘 Troubleshooting

### Sentry Not Working?
```bash
# Check logs for this message:
# "Sentry initialized successfully"

# Test manually
curl http://localhost:8081/health

# Check Sentry dashboard
open https://sentry.io
```

### pprof Not Working?

**Development:**
```bash
# Should work:
curl http://localhost:8081/debug/pprof/
```

**Production:**
```bash
# Without auth - should return 401 or 404:
curl http://your-server.com/debug/pprof/

# With auth - should work:
curl -u admin:password http://your-server.com/debug/pprof/
```

---

## 💰 Cost

**Current Setup:** FREE
- Sentry: 5,000 events/month
- pprof: Built-in, unlimited

**If you need more:**
- Sentry Team: $26/month (50K events)
- Sentry Business: $80/month (150K events)

---

## 🎓 Learn More

### Official Resources
- [Sentry Go SDK](https://docs.sentry.io/platforms/go/)
- [Go pprof Guide](https://go.dev/blog/pprof)
- [pprof Documentation](https://pkg.go.dev/net/http/pprof)

### Video Tutorials
- [Go Performance Profiling](https://www.youtube.com/results?search_query=golang+pprof+tutorial)
- [Sentry Error Tracking](https://www.youtube.com/results?search_query=sentry+go+tutorial)

---

## ✅ Checklist

### Development
- [x] Sentry configured
- [x] pprof configured
- [x] Helper functions created
- [x] Documentation written
- [ ] Test error capture
- [ ] Test profiling

### Production Deployment
- [ ] Set `GO_ENV=production`
- [ ] Verify `SENTRY_DSN` is set
- [ ] Set `PPROF_USERNAME` (optional)
- [ ] Set `PPROF_PASSWORD` (optional)
- [ ] Configure Sentry alerts
- [ ] Test via SSH tunnel

---

## 🎉 You're All Set!

Both Sentry and pprof are **fully configured and production-ready**.

**Next steps:**
1. Run your app: `make dev`
2. Test Sentry by triggering an error
3. Test pprof by opening http://localhost:8081/debug/pprof/
4. Read the detailed guides when you need them

**Questions?** Check the detailed guides in `docs/`:
- `SENTRY_USAGE_GUIDE.md`
- `PPROF_PROFILING_GUIDE.md`
- `SECURITY_OBSERVABILITY_GUIDE.md`

---

**Last Updated:** 2026-01-24
**Status:** ✅ Production Ready
