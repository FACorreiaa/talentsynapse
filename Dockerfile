# =========================================================================
# Stage 1: Builder
# =========================================================================
FROM golang:1.23-alpine AS builder

# Install build dependencies including Node.js
RUN apk add --no-cache git curl bash nodejs npm

WORKDIR /app

# Install Go tools
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate Templ templates
RUN templ generate

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# =========================================================================
# Stage 2: Runner
# =========================================================================
FROM alpine:3.19 AS runner

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy binary and assets from builder
COPY --from=builder /app/server .
COPY --from=builder /app/assets ./assets

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
CMD ["./server"]
