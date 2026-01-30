package middleware

import (
	"net/http"

	"github.com/FACorreiaa/talentsynapse/internal/observability"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// NewRelicMiddleware instruments HTTP handlers with New Relic APM
func NewRelicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app := observability.GetNewRelicApp()
		if app == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Create transaction name from method and path
		txnName := r.Method + " " + r.URL.Path
		txn := app.StartTransaction(txnName)
		defer txn.End()

		// Add request attributes
		txn.SetWebRequestHTTP(r)

		// Wrap the response writer to capture status code
		w = txn.SetWebResponse(w)

		// Add the transaction to the request context
		r = newrelic.RequestWithTransactionContext(r, txn)

		next.ServeHTTP(w, r)
	})
}

// WrapHandler wraps a specific handler function with New Relic transaction
func WrapHandler(pattern string, handler http.HandlerFunc) http.HandlerFunc {
	app := observability.GetNewRelicApp()
	if app == nil {
		return handler
	}

	_, wrappedHandler := newrelic.WrapHandleFunc(app, pattern, handler)
	return wrappedHandler
}
