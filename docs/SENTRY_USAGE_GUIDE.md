# Sentry Integration Guide

This guide explains how to use Sentry error tracking in the SkillSphere application.

## Configuration

Sentry is initialized in `cmd/server/main.go` using the `SENTRY_DSN` environment variable from `.env`:

```go
sentry.Init(sentry.ClientOptions{
    Dsn:              os.Getenv("SENTRY_DSN"),
    Environment:      os.Getenv("GO_ENV"),
    TracesSampleRate: 1.0,
    Debug:            os.Getenv("GO_ENV") == "development",
})
```

The HTTP middleware is added early in the middleware chain in `internal/server/routes.go` to capture all HTTP-related errors and panics.

## Basic Usage

### 1. Capturing Errors

Use the helper functions in `internal/server/sentry_helpers.go`:

```go
func (h *Handler) DoSomething(w http.ResponseWriter, r *http.Request) {
    result, err := h.service.PerformOperation()
    if err != nil {
        // This will send the error to Sentry
        CaptureError(r, err)
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }

    // Handle success...
}
```

### 2. Capturing Messages

```go
func (h *Handler) SuspiciousActivity(w http.ResponseWriter, r *http.Request) {
    // Log a message to Sentry for monitoring
    CaptureMessage(r, "User attempted unauthorized access")

    http.Error(w, "Forbidden", http.StatusForbidden)
}
```

### 3. Adding Context to Events

Use `WithSentryScope` to add extra information:

```go
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
    userID := auth.GetUserIDFromSession(r)

    // Add context that will be attached to all Sentry events in this request
    WithSentryScope(r, func(scope *sentry.Scope) {
        scope.SetUser(sentry.User{
            ID:    userID,
        })
        scope.SetTag("operation", "profile_update")
        scope.SetExtra("profile_data", map[string]interface{}{
            "fields_updated": []string{"name", "bio"},
        })
    })

    if err := h.service.UpdateProfile(userID, data); err != nil {
        CaptureError(r, err)
        http.Error(w, "Failed to update profile", http.StatusInternalServerError)
        return
    }

    // Success response...
}
```

### 4. Using Sentry Hub Directly

For more advanced use cases, you can access the Sentry hub from the request context:

```go
func (h *Handler) ComplexOperation(w http.ResponseWriter, r *http.Request) {
    if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
        // Add breadcrumbs to track the flow
        hub.AddBreadcrumb(&sentry.Breadcrumb{
            Type:     "info",
            Category: "operation",
            Message:  "Starting complex operation",
            Level:    sentry.LevelInfo,
        }, nil)

        // Perform operation...

        hub.AddBreadcrumb(&sentry.Breadcrumb{
            Type:     "info",
            Category: "operation",
            Message:  "Operation completed successfully",
            Level:    sentry.LevelInfo,
        }, nil)
    }
}
```

## What Gets Captured Automatically

The Sentry HTTP middleware automatically captures:

1. **Panics**: Any panic in your handlers will be caught and sent to Sentry
2. **Request Context**: HTTP method, URL, headers, query parameters
3. **User IP**: Client IP address (respecting proxies)
4. **Stack Traces**: Full stack trace when errors occur

## Configuration Options

The middleware is configured in `routes.go` with these options:

```go
sentryhttp.Options{
    Repanic:         true,  // Re-throw panic after capturing (for graceful handling)
    WaitForDelivery: false, // Don't block request waiting for Sentry
    Timeout:         2 * time.Second,
}
```

## Best Practices

1. **Always capture errors**: Whenever you log an error, consider if it should go to Sentry
2. **Add context**: Use `WithSentryScope` to add user IDs, operation types, etc.
3. **Use breadcrumbs**: Track the flow of operations to debug issues
4. **Don't capture expected errors**: Only send errors that indicate problems
5. **Sensitive data**: Be careful not to send passwords, tokens, or PII to Sentry

## Testing

To test Sentry integration locally:

1. Ensure `SENTRY_DSN` is set in `.env`
2. Set `GO_ENV=development` to enable debug mode
3. Trigger an error in your application
4. Check the Sentry dashboard at: https://sentry.io

## Environment Variables

- `SENTRY_DSN`: Your Sentry project DSN (required)
- `GO_ENV`: Environment name (development/production)

## Resources

- [Sentry Go SDK Documentation](https://docs.sentry.io/platforms/go/)
- [Sentry HTTP Integration](https://docs.sentry.io/platforms/go/guides/http/)
