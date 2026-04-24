package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/officialasishkumar/senior-systems-lab/internal/broker"
	"github.com/officialasishkumar/senior-systems-lab/internal/observability"
)

type HTTP struct {
	server  *http.Server
	broker  *broker.Broker
	metrics *observability.Metrics
	logger  *slog.Logger
}

func NewHTTP(addr string, b *broker.Broker, metrics *observability.Metrics, logger *slog.Logger) *HTTP {
	h := &HTTP{broker: b, metrics: metrics, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /readyz", h.ready)
	mux.HandleFunc("GET /metrics", h.prometheus)
	mux.HandleFunc("GET /queue/stats", h.queueStats)
	mux.HandleFunc("GET /queue/dead-letters", h.deadLetters)
	mux.HandleFunc("POST /queue/publish", h.publish)
	h.server = &http.Server{
		Addr:              addr,
		Handler:           observability.TraceMiddleware(h.requestLog(mux)),
		ReadHeaderTimeout: 3 * time.Second,
	}
	return h
}

func (h *HTTP) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() { errCh <- h.server.ListenAndServe() }()
	select {
	case <-ctx.Done():
		return h.server.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

func (h *HTTP) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

func (h *HTTP) requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.metrics.IncHTTPRequests()
		next.ServeHTTP(w, r)
		h.metrics.ObserveLatency("http", time.Since(start))
		h.logger.InfoContext(r.Context(), "http request", "method", r.Method, "path", r.URL.Path, "trace_id", observability.TraceID(r.Context()))
	})
}

func (h *HTTP) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HTTP) ready(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *HTTP) prometheus(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	h.metrics.WritePrometheus(w)
}

func (h *HTTP) queueStats(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.broker.Stats())
}

func (h *HTTP) deadLetters(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.broker.DeadLetters())
}

func (h *HTTP) publish(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic   string            `json:"topic"`
		Payload string            `json:"payload"`
		Headers map[string]string `json:"headers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.Topic == "" || req.Payload == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "topic and payload are required"})
		return
	}
	err := h.broker.Publish(r.Context(), broker.Message{
		Topic:   req.Topic,
		Payload: req.Payload,
		TraceID: observability.TraceID(r.Context()),
		Headers: req.Headers,
	})
	if errors.Is(err, broker.ErrQueueFull) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "queue full"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted", "trace_id": observability.TraceID(r.Context())})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
