package observability

import (
	"context"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// StartDatastoreSegment creates a New Relic datastore segment for database operations
// Usage:
//
//	segment := observability.StartDatastoreSegment(ctx, "SELECT", "users", "SELECT * FROM users WHERE id = $1")
//	defer segment.End()
//	// ... execute query ...
func StartDatastoreSegment(ctx context.Context, operation, collection, query string) *newrelic.DatastoreSegment {
	txn := newrelic.FromContext(ctx)
	if txn == nil {
		return &newrelic.DatastoreSegment{} // Return empty segment if no transaction
	}

	return &newrelic.DatastoreSegment{
		StartTime:          txn.StartSegmentNow(),
		Product:            newrelic.DatastorePostgres,
		Collection:         collection, // table name
		Operation:          operation,  // SELECT, INSERT, UPDATE, DELETE
		ParameterizedQuery: query,
	}
}

// StartExternalSegment creates a New Relic segment for external HTTP calls
// Usage:
//
//	segment := observability.StartExternalSegment(ctx, "https://api.example.com/users")
//	defer segment.End()
//	// ... make HTTP request ...
func StartExternalSegment(ctx context.Context, url string) *newrelic.ExternalSegment {
	txn := newrelic.FromContext(ctx)
	if txn == nil {
		return &newrelic.ExternalSegment{} // Return empty segment if no transaction
	}

	return &newrelic.ExternalSegment{
		StartTime: txn.StartSegmentNow(),
		URL:       url,
	}
}

// StartSegment creates a generic New Relic segment for custom operations
// Usage:
//
//	segment := observability.StartSegment(ctx, "process-image")
//	defer segment.End()
//	// ... process image ...
func StartSegment(ctx context.Context, name string) *newrelic.Segment {
	txn := newrelic.FromContext(ctx)
	if txn == nil {
		return &newrelic.Segment{} // Return empty segment if no transaction
	}

	return txn.StartSegment(name)
}

// NoticeError records an error in the current New Relic transaction
func NoticeError(ctx context.Context, err error) {
	txn := newrelic.FromContext(ctx)
	if txn != nil && err != nil {
		txn.NoticeError(err)
	}
}

// AddAttribute adds a custom attribute to the current transaction
func AddAttribute(ctx context.Context, key string, value interface{}) {
	txn := newrelic.FromContext(ctx)
	if txn != nil {
		txn.AddAttribute(key, value)
	}
}
