package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTraceMiddlewarePropagatesHeader(t *testing.T) {
	handler := TraceMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := TraceID(r.Context()); got != "trace-test" {
			t.Fatalf("trace id = %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderTraceID, "trace-test")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(HeaderTraceID); got != "trace-test" {
		t.Fatalf("response trace header = %q", got)
	}
}
