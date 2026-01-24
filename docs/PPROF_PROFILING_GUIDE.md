# Go pprof Profiling Guide

This guide explains how to use Go's built-in pprof profiler to analyze your application's performance in both development and production environments.

## Overview

pprof is a powerful profiling tool built into Go that helps you:
- Identify CPU bottlenecks
- Find memory leaks
- Analyze goroutine usage
- Detect lock contention
- Profile blocking operations

## Security Configuration

### Development Mode
In development (`GO_ENV=development`), pprof endpoints are **open and unrestricted** at:
```
http://localhost:8081/debug/pprof/
```

### Production Mode
In production, pprof has **multiple layers of security**:

1. **Authentication Required**: Must provide username/password via Basic Auth
2. **Disabled by Default**: Set `PPROF_ENABLE=true` to enable
3. **Credentials Required**: Must set `PPROF_USERNAME` and `PPROF_PASSWORD`

#### Environment Variables

Add to your `.env` or production environment:

```bash
# Enable pprof in production
PPROF_ENABLE=true

# Set strong credentials
PPROF_USERNAME=admin
PPROF_PASSWORD=YourVerySecurePassword123!
```

**Security Best Practices:**
- Use a strong, unique password
- Never commit credentials to version control
- Consider restricting to localhost only (SSH tunnel required)
- Use temporary credentials and rotate them regularly

## Available Endpoints

All endpoints are under `/debug/pprof/`:

| Endpoint | Description |
|----------|-------------|
| `/debug/pprof/` | Index page with links to all profiles |
| `/debug/pprof/heap` | Memory allocation profile |
| `/debug/pprof/goroutine` | Stack traces of all current goroutines |
| `/debug/pprof/threadcreate` | Stack traces that led to thread creation |
| `/debug/pprof/block` | Stack traces that led to blocking on synchronization primitives |
| `/debug/pprof/mutex` | Stack traces of holders of contended mutexes |
| `/debug/pprof/profile` | CPU profile (30s default) |
| `/debug/pprof/trace` | Execution trace |
| `/debug/pprof/allocs` | All past memory allocations |
| `/debug/pprof/cmdline` | Command line invocation |

## Usage Examples

### 1. Development (Local)

#### View in Browser
```bash
# Open the index page
open http://localhost:8081/debug/pprof/

# View specific profiles
open http://localhost:8081/debug/pprof/heap
open http://localhost:8081/debug/pprof/goroutine
```

#### CPU Profile (30 seconds)
```bash
go tool pprof http://localhost:8081/debug/pprof/profile?seconds=30
```

#### Memory Profile
```bash
go tool pprof http://localhost:8081/debug/pprof/heap
```

#### Goroutine Profile
```bash
go tool pprof http://localhost:8081/debug/pprof/goroutine
```

### 2. Production (with Authentication)

#### Using curl with Basic Auth
```bash
# CPU profile
curl -u admin:password http://your-server.com/debug/pprof/profile?seconds=30 -o cpu.prof

# Heap profile
curl -u admin:password http://your-server.com/debug/pprof/heap -o heap.prof

# Analyze locally
go tool pprof cpu.prof
```

#### Using go tool pprof with Auth
```bash
# Set credentials in URL (not recommended for shared terminals)
go tool pprof http://admin:password@your-server.com/debug/pprof/profile

# Better: Use SSH tunnel (no credentials in URL)
ssh -L 6060:localhost:8081 user@your-server.com
go tool pprof http://localhost:6060/debug/pprof/profile
```

#### SSH Tunnel Method (Most Secure)
```bash
# 1. Create SSH tunnel
ssh -L 6060:localhost:8081 user@your-production-server.com

# 2. In another terminal, profile via tunnel
go tool pprof http://admin:password@localhost:6060/debug/pprof/profile?seconds=30

# 3. Analyze
(pprof) top10
(pprof) web
```

## Analyzing Profiles

### Interactive Commands

Once you've loaded a profile with `go tool pprof`, you can use:

```bash
(pprof) top10          # Show top 10 functions
(pprof) top20 -cum     # Show top 20 by cumulative time
(pprof) list funcName  # Show source code for function
(pprof) web            # Generate and open SVG graph (requires graphviz)
(pprof) png            # Generate PNG graph
(pprof) pdf            # Generate PDF report
(pprof) peek funcName  # Show callers and callees
```

### Web UI

Modern pprof has a web UI:

```bash
# Start web server on :8080
go tool pprof -http=:8080 http://localhost:8081/debug/pprof/profile?seconds=30
```

This opens an interactive visualization in your browser with:
- Flame graphs
- Call graphs
- Top functions
- Source code view

### Common Profiling Scenarios

#### Finding CPU Hotspots
```bash
# Capture 30-second CPU profile
go tool pprof -http=:8080 http://localhost:8081/debug/pprof/profile?seconds=30

# Look at flame graph to see where CPU time is spent
```

#### Finding Memory Leaks
```bash
# Take heap snapshot
go tool pprof -http=:8080 http://localhost:8081/debug/pprof/heap

# Compare two heap snapshots
go tool pprof -base heap1.prof heap2.prof

# Look for:
# - Increasing allocations over time
# - Unexpected memory retention
# - Large objects not being freed
```

#### Analyzing Goroutine Leaks
```bash
# View current goroutines
go tool pprof http://localhost:8081/debug/pprof/goroutine

# Commands:
(pprof) top          # See which goroutines are most common
(pprof) traces       # See stack traces
```

#### Finding Lock Contention
```bash
# Must enable mutex profiling in your app first
# Add to main.go:
# runtime.SetMutexProfileFraction(1)

go tool pprof http://localhost:8081/debug/pprof/mutex

# Look for:
# - Functions spending time waiting for locks
# - Hot locks that multiple goroutines compete for
```

## Continuous Profiling

For continuous profiling in production, you can:

1. **Periodic Snapshots**: Use cron to capture profiles regularly
```bash
#!/bin/bash
# Save to /var/profiling/$(date +%Y%m%d-%H%M%S)-cpu.prof
curl -u admin:password http://localhost:8081/debug/pprof/profile?seconds=30 \
  -o /var/profiling/$(date +%Y%m%d-%H%M%S)-cpu.prof
```

2. **Use Monitoring Tools**:
   - Datadog
   - New Relic
   - Grafana Pyroscope
   - Elastic APM

## Enabling Additional Profiles

### Block Profile
Add to your `main.go`:
```go
import "runtime"

func main() {
    // Enable block profiling
    runtime.SetBlockProfileRate(1)

    // Rest of your code...
}
```

### Mutex Profile
Add to your `main.go`:
```go
import "runtime"

func main() {
    // Enable mutex profiling
    runtime.SetMutexProfileFraction(1)

    // Rest of your code...
}
```

## Tips and Best Practices

### Development
- Profile regularly during development to catch issues early
- Use the `-http` flag for easier visualization
- Compare before/after profiles when optimizing

### Production
- Keep pprof disabled by default
- Enable only when investigating issues
- Use strong authentication
- Prefer SSH tunnels over exposing to internet
- Profile during low-traffic periods if possible
- CPU profiling adds ~5% overhead for the duration

### Interpreting Results
- Focus on cumulative time for CPU profiles
- Look for unexpected allocations in heap profiles
- Check for goroutine leaks in long-running services
- Memory profiles show allocated memory, not necessarily used memory

## Troubleshooting

### "connection refused"
- Check server is running
- Verify port number
- Check firewall rules

### "401 Unauthorized"
- Verify `PPROF_USERNAME` and `PPROF_PASSWORD` are set
- Check credentials in request
- Ensure `PPROF_ENABLE=true` in production

### "404 Not Found"
- Check `GO_ENV` setting
- Verify `PPROF_ENABLE=true` for production
- Ensure credentials are configured

### Profile shows no data
- Increase profiling duration: `?seconds=60`
- Ensure application is under load
- Check that profiling type is appropriate for issue

## Resources

- [Official pprof Documentation](https://pkg.go.dev/net/http/pprof)
- [Go Blog: Profiling Go Programs](https://go.dev/blog/pprof)
- [Practical Go: Real World Advice for Writing Maintainable Go Programs](https://dave.cheney.net/practical-go/presentations/qcon-china.html)
- [pprof User Guide](https://github.com/google/pprof/blob/master/doc/README.md)

## Quick Reference

```bash
# CPU Profile (30s)
go tool pprof http://localhost:8081/debug/pprof/profile?seconds=30

# Heap Profile
go tool pprof http://localhost:8081/debug/pprof/heap

# Goroutine Profile
go tool pprof http://localhost:8081/debug/pprof/goroutine

# Web UI (most user-friendly)
go tool pprof -http=:8080 http://localhost:8081/debug/pprof/profile?seconds=30

# Production (via SSH tunnel)
ssh -L 6060:localhost:8081 user@server
go tool pprof -http=:8080 http://admin:pass@localhost:6060/debug/pprof/profile?seconds=30
```
