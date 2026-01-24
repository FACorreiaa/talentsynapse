# pprof Quick Start Guide

## Development Usage (Local)

When running locally with `GO_ENV=development`, pprof endpoints are **open and unrestricted**.

### View Available Profiles

```bash
# Open in browser
open http://localhost:8081/debug/pprof/

# Or with curl
curl http://localhost:8081/debug/pprof/
```

### Common Profiling Commands

#### 1. CPU Profile (30 seconds)
```bash
go tool pprof http://localhost:8081/debug/pprof/profile?seconds=30

# With interactive web UI (recommended)
go tool pprof -http=:8080 http://localhost:8081/debug/pprof/profile?seconds=30
```

#### 2. Memory (Heap) Profile
```bash
go tool pprof http://localhost:8081/debug/pprof/heap

# With web UI
go tool pprof -http=:8080 http://localhost:8081/debug/pprof/heap
```

#### 3. Goroutine Profile
```bash
go tool pprof http://localhost:8081/debug/pprof/goroutine

# Or view in browser
open http://localhost:8081/debug/pprof/goroutine
```

#### 4. All Allocations
```bash
go tool pprof http://localhost:8081/debug/pprof/allocs
```

## Production Usage

### Prerequisites

Set these environment variables:

```bash
PPROF_USERNAME=your-username
PPROF_PASSWORD=your-secure-password
```

### Option 1: Direct Access (with Basic Auth)

```bash
# CPU profile
go tool pprof http://username:password@your-server.com/debug/pprof/profile?seconds=30

# Heap profile
go tool pprof http://username:password@your-server.com/debug/pprof/heap
```

### Option 2: SSH Tunnel (Most Secure - Recommended)

```bash
# 1. Create SSH tunnel from your local machine
ssh -L 6060:localhost:8081 user@your-production-server.com

# 2. In another terminal, profile via the tunnel
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30

# The browser will open with an interactive UI
```

### Option 3: Save Profile and Analyze Locally

```bash
# On production server (or via curl)
curl -u username:password http://localhost:8081/debug/pprof/profile?seconds=30 -o cpu.prof
curl -u username:password http://localhost:8081/debug/pprof/heap -o heap.prof

# Copy to your local machine
scp user@server:/path/to/cpu.prof .

# Analyze locally
go tool pprof -http=:8080 cpu.prof
```

## Common Scenarios

### Finding CPU Bottlenecks

```bash
# Capture 30s of CPU activity
go tool pprof -http=:8080 http://localhost:8081/debug/pprof/profile?seconds=30

# In the web UI:
# 1. Click "Flame Graph" - see where time is spent
# 2. Click "Top" - see top functions by CPU time
# 3. Click "Source" - see actual code
```

### Finding Memory Leaks

```bash
# Take initial snapshot
curl http://localhost:8081/debug/pprof/heap -o heap1.prof

# Wait some time (let app run)
sleep 300

# Take second snapshot
curl http://localhost:8081/debug/pprof/heap -o heap2.prof

# Compare
go tool pprof -base heap1.prof heap2.prof
```

### Finding Goroutine Leaks

```bash
# View current goroutines
go tool pprof http://localhost:8081/debug/pprof/goroutine

# Commands:
(pprof) top       # Most common goroutines
(pprof) list      # Show source code
(pprof) traces    # Show stack traces
```

## Interactive Commands

Once in pprof interactive mode:

```bash
(pprof) top           # Show top 10 entries
(pprof) top20         # Show top 20 entries
(pprof) top -cum      # Sort by cumulative time
(pprof) list funcName # Show annotated source for function
(pprof) web           # Generate SVG graph (requires graphviz)
(pprof) pdf           # Generate PDF report
(pprof) png           # Generate PNG image
(pprof) help          # Show all commands
```

## Security Notes

### Development
- No authentication required
- All endpoints accessible at `/debug/pprof/*`
- Safe for local development only

### Production
- **Basic Authentication required** (username + password)
- Set `PPROF_USERNAME` and `PPROF_PASSWORD` environment variables
- If not set, pprof returns 404 (disabled)
- Use SSH tunnels to avoid exposing credentials

## Environment Variables

```bash
# Development (pprof open)
GO_ENV=development

# Production (pprof requires auth)
GO_ENV=production
PPROF_USERNAME=admin
PPROF_PASSWORD=your-secure-password-here
```

## Tips

1. **Use the web UI** (`-http=:8080` flag) - it's much easier than command line
2. **Profile during load** - CPU profiles are most useful when the app is under load
3. **Take snapshots** - Memory profiles are best analyzed by comparing snapshots over time
4. **Don't profile too long** - 30s is usually enough for CPU, shorter for others
5. **Install graphviz** for better visualizations: `brew install graphviz` (macOS)

## Troubleshooting

### "connection refused"
- Server not running
- Wrong port number
- Check firewall

### "401 Unauthorized" (Production)
- Set `PPROF_USERNAME` and `PPROF_PASSWORD`
- Check credentials in request
- Verify environment variables are exported

### "404 Not Found" (Production)
- `PPROF_USERNAME` or `PPROF_PASSWORD` not set
- pprof is intentionally disabled (security feature)

## Resources

- [pprof Documentation](https://pkg.go.dev/net/http/pprof)
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [Full Guide](./PPROF_PROFILING_GUIDE.md)
