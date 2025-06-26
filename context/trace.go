package context

import "context"

type traceContextKey struct{}

var TraceContextKey = traceContextKey{}

type Trace struct {
	TraceID   string
	SpanID    string
	TraceTrue bool
}

func SetTrace(ctx context.Context, traceID, spanID string, traceTrue bool) context.Context {
	return context.WithValue(ctx, TraceContextKey, Trace{
		TraceID:   traceID,
		SpanID:    spanID,
		TraceTrue: traceTrue,
	})
}

func GetTrace(ctx context.Context) (Trace, bool) {
	if v := ctx.Value(TraceContextKey); v != nil {
		return v.(Trace), true
	}

	return Trace{}, false
}
