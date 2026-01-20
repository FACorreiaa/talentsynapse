# Build Optimization Guide

This document outlines all optimizations implemented for bundling everything into the Go binary.

## Current Build Stats

### Before Optimization
- **Binary Size**: 8.9 MB
- **CSS Size**: 225 KB (minified)
- **Total Assets**: ~576 KB

### Optimization Goals
1. Minimize CSS bundle size
2. Optimize binary size
3. Enable asset compression
4. Tree-shake unused code

---

## 1. CSS Optimization

### Tailwind Configuration
The `tailwind.config.js` has been optimized:

```javascript
daisyui: {
    themes: ["dark"],        // Single theme (was: ["light", "dark"])
    darkTheme: "dark",
    base: true,
    styled: true,
    utils: true,
    logs: false,             // Disable console logs
}
```

**Improvements**:
- ✅ CSS minification enabled via `--minify` flag
- ✅ Single theme reduces CSS by ~10-15%
- ✅ Tailwind v4 automatically tree-shakes unused utilities
- ✅ DaisyUI only bundles components used in `content` paths

### Build Command
```bash
tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --minify
```

---

## 2. Go Binary Optimization

### Build Flags
Current build command in `Makefile`:

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/server ./cmd/server
```

**Flags Explained**:
- `CGO_ENABLED=0`: Pure Go binary (no C dependencies)
- `-s`: Strip symbol table
- `-w`: Strip DWARF debugging info
- Combined `-s -w` reduces binary ~20-30%

### Additional Optimization Options

#### 1. UPX Compression (Optional)
Install UPX and compress the binary:

```bash
# macOS
brew install upx

# Compress binary (reduces 50-70% but adds startup overhead)
upx --best --lzma ./bin/server

# Conservative compression (30-40% reduction, minimal overhead)
upx -7 ./bin/server
```

**Trade-offs**:
- ✅ Significantly smaller binary size
- ⚠️ Slight startup time increase (~50-100ms)
- ⚠️ Some security scanners may flag compressed binaries

#### 2. Optimize Embedded Assets

Currently using `embed.FS` in `assets/efs.go`. All assets are embedded.

**Asset Compression Script**:

```bash
#!/bin/bash
# scripts/compress-assets.sh

# Minify JS files (if any custom JS added)
# npm install -g terser
find assets/js -name "*.js" ! -name "*.min.js" -type f -exec sh -c \
  'terser "$1" -c -m -o "${1%.js}.min.js"' _ {} \;

# Optimize images (if any added later)
# brew install imageoptim-cli
# imageoptim --imagealpha --jpegmini assets/static/*.{png,jpg,jpeg}

# Pre-compress assets with Brotli and Gzip for serving
find assets -type f \( -name "*.css" -o -name "*.js" -o -name "*.json" \) -exec sh -c \
  'gzip -9 -k "$1" && brotli -q 11 -k "$1"' _ {} \;
```

---

## 3. Runtime Compression

### Middleware Compression
The app already uses compression middleware in `routes.go`:

```go
r.Use(middleware.Compress(5)) // Gzip level 5
```

**Enhancement**: Serve pre-compressed assets:

```go
// In routes.go, add before file server
r.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check if client accepts Brotli
        if strings.Contains(r.Header.Get("Accept-Encoding"), "br") {
            brPath := r.URL.Path + ".br"
            if _, err := assets.Files.Open(strings.TrimPrefix(brPath, "/")); err == nil {
                w.Header().Set("Content-Encoding", "br")
                w.Header().Set("Content-Type", getContentType(r.URL.Path))
                r.URL.Path = brPath
            }
        }
        next.ServeHTTP(w, r)
    })
})
```

---

## 4. Database Configuration

### Connection Pool Optimization
Current settings in `internal/database/database.go`:

```go
config.MaxConns = 25         // Maximum concurrent connections
config.MinConns = 5          // Minimum idle connections
config.MaxConnLifetime = time.Hour
config.MaxConnIdleTime = 30 * time.Minute
config.HealthCheckPeriod = time.Minute
```

**For deployment**:
- Small instance: `MaxConns = 10-15`
- Medium instance: `MaxConns = 25-50`
- Large instance: `MaxConns = 50-100`

---

## 5. Production Build Script

Create `scripts/build-prod.sh`:

```bash
#!/bin/bash
set -e

echo "🧹 Cleaning previous builds..."
rm -rf bin/ tmp/
rm -f assets/css/output.css

echo "📝 Generating Templ templates..."
templ generate

echo "🎨 Building optimized CSS..."
tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --minify

echo "🔨 Building Go binary..."
CGO_ENABLED=0 go build \
  -ldflags="-s -w -X main.version=$(git describe --tags --always)" \
  -o ./bin/server \
  ./cmd/server

echo "📊 Build Statistics:"
ls -lh bin/server | awk '{print "Binary size:", $5}'
ls -lh assets/css/output.css | awk '{print "CSS size:", $5}'

# Optional: UPX compression
# echo "🗜️  Compressing binary with UPX..."
# upx -7 ./bin/server

echo "✅ Production build complete!"
```

Make it executable:
```bash
chmod +x scripts/build-prod.sh
```

---

## 6. Deployment Checklist

### Before Deploying:

1. **Environment Variables**:
   ```bash
   GO_ENV=production
   DATABASE_URL=postgres://...
   PORT=8080
   ```

2. **Build Production Binary**:
   ```bash
   ./scripts/build-prod.sh
   ```

3. **Test Binary**:
   ```bash
   ./bin/server
   ```

4. **Check Binary Size**:
   ```bash
   ls -lh bin/server
   ```

5. **Verify Assets Are Embedded**:
   ```bash
   # Binary should work without assets/ directory
   mkdir test-deploy && cp bin/server test-deploy/
   cd test-deploy && ./server
   ```

### Docker Optimization

For Docker deployments, use multi-stage builds:

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git tailwindcss upx

WORKDIR /app
COPY .. .

RUN templ generate
RUN tailwindcss -i ./assets/css/input.css -o ./assets/css/output.css --minify
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server
RUN upx -7 server

# Runtime stage
FROM scratch
COPY --from=builder /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
```

**Result**: Final image size ~5-10 MB (instead of 100+ MB)

---

## 7. Monitoring Bundle Size

Add to `Makefile`:

```makefile
size-report: ## Show size breakdown of build artifacts
	@echo "📊 Build Size Report"
	@echo "===================="
	@echo ""
	@echo "Binary:"
	@ls -lh bin/server 2>/dev/null | awk '{print "  Size:", $$5}' || echo "  Not built"
	@echo ""
	@echo "CSS:"
	@ls -lh assets/css/output.css 2>/dev/null | awk '{print "  Size:", $$5}' || echo "  Not built"
	@echo ""
	@echo "Total Assets:"
	@du -sh assets/ 2>/dev/null | awk '{print "  Size:", $$1}' || echo "  Not found"
	@echo ""
	@echo "Embedded in binary:"
	@du -sh bin/ 2>/dev/null | awk '{print "  Total:", $$1}' || echo "  Not built"
```

---

## 8. Expected Results

### Optimized Build Sizes:

| Component | Before | After Optimizations | Savings |
|-----------|--------|-------------------|---------|
| Go Binary | 8.9 MB | 8.9 MB (6-7 MB with UPX) | 0-30% |
| CSS       | 225 KB | 205-215 KB | ~5-10% |
| Total Binary | 8.9 MB | 6-7 MB (with UPX) | ~25-30% |

### Additional Optimizations Applied:
- ✅ Single DaisyUI theme
- ✅ CSS minification
- ✅ Binary stripping (-s -w)
- ✅ CGO disabled
- ✅ Gzip compression middleware
- ✅ Connection pool tuning

### Future Optimizations:
- Pre-compress assets with Brotli
- Implement asset fingerprinting/versioning
- Add cache headers for static assets
- Consider lazy-loading non-critical CSS
- Split CSS into critical/non-critical paths

---

## 9. Verification Commands

```bash
# Check binary is statically linked
file bin/server
# Should output: "statically linked"

# Check binary symbols are stripped
nm bin/server 2>&1
# Should output: "no symbols"

# Test binary works without assets directory
mkdir /tmp/test && cp bin/server /tmp/test/ && cd /tmp/test && ./server

# Check embedded files
go list -f '{{.EmbedFiles}}' ./assets
```

---

## Summary

Your current setup is already well-optimized for production deployment:

1. ✅ All assets embedded in binary via `embed.FS`
2. ✅ CSS minified with Tailwind
3. ✅ Binary stripped of debug symbols
4. ✅ Gzip compression enabled
5. ✅ Single-theme DaisyUI configuration

The binary is fully self-contained and can be deployed as a single file with just environment variables needed for configuration.

**Deploy command**:
```bash
# Copy binary to server
scp bin/server user@server:/app/

# Run on server
DATABASE_URL=postgres://... PORT=8080 /app/server
```
