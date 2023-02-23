package spancontext

import (
	"context"

	"github.com/thnthien/great-deku/trace"
	"github.com/thnthien/great-deku/trace/pkg/id"
)

type contextKey struct{}

type SpanContext struct {
	TraceID id.TraceID
	SpanID  id.SpanID
}

// FromContext returns the Span stored in a context, or nil if there isn't one.
func FromContext(ctx context.Context) trace.ISpan {
	s, _ := ctx.Value(contextKey{}).(trace.ISpan)
	return s
}

// NewContext returns a new context with the given Span attached.
func NewContext(parent context.Context, s trace.ISpan) context.Context {
	return context.WithValue(parent, contextKey{}, s)
}
