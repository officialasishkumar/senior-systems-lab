package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type traceKey struct{}

const HeaderTraceID = "X-Trace-ID"

func WithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		traceID = NewTraceID()
	}
	return context.WithValue(ctx, traceKey{}, traceID)
}

func TraceID(ctx context.Context) string {
	value, _ := ctx.Value(traceKey{}).(string)
	return value
}

func NewTraceID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "trace-unavailable"
	}
	return hex.EncodeToString(b[:])
}

func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get(HeaderTraceID)
		ctx := WithTraceID(r.Context(), traceID)
		w.Header().Set(HeaderTraceID, TraceID(ctx))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
