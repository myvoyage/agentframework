package telemetry

import (
	"context"
)

// TracePlaceholder is a minimal wrapper to return a no-op trace span.
func TraceFromContext(ctx context.Context) context.Context {
	return ctx
}

func EndTrace(ctx context.Context) {}
