package server

import (
	"net/http"

	"github.com/getsentry/sentry-go"
)

// CaptureError captures an error to Sentry with context from the HTTP request
func CaptureError(r *http.Request, err error) {
	if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
		hub.CaptureException(err)
	}
}

// CaptureMessage captures a message to Sentry with context from the HTTP request
func CaptureMessage(r *http.Request, message string) {
	if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
		hub.CaptureMessage(message)
	}
}

// WithSentryScope executes a function with a configured Sentry scope
// This allows you to add extra context, tags, or breadcrumbs to events
func WithSentryScope(r *http.Request, fn func(scope *sentry.Scope)) {
	if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
		hub.WithScope(fn)
	}
}

// Example usage in a handler:
//
// func (s *Server) handleSomething(w http.ResponseWriter, r *http.Request) {
//     userID := getUserIDFromSession(r)
//
//     // Add context to all Sentry events in this request
//     WithSentryScope(r, func(scope *sentry.Scope) {
//         scope.SetTag("user_id", userID)
//         scope.SetContext("request_info", map[string]interface{}{
//             "path": r.URL.Path,
//             "method": r.Method,
//         })
//     })
//
//     // If an error occurs
//     if err := doSomething(); err != nil {
//         CaptureError(r, err)
//         http.Error(w, "Internal error", http.StatusInternalServerError)
//         return
//     }
//
//     // Or capture a custom message
//     CaptureMessage(r, "User performed suspicious action")
// }
