package server

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/officialasishkumar/pulsemesh/internal/broker"
	"github.com/officialasishkumar/pulsemesh/internal/observability"
)

type TCP struct {
	addr     string
	broker   *broker.Broker
	metrics  *observability.Metrics
	logger   *slog.Logger
	listener net.Listener
	mu       sync.Mutex
}

func NewTCP(addr string, b *broker.Broker, metrics *observability.Metrics, logger *slog.Logger) *TCP {
	return &TCP{addr: addr, broker: b, metrics: metrics, logger: logger}
}

func (s *TCP) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.listener = ln
	s.mu.Unlock()
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || ctx.Err() != nil {
				return context.Canceled
			}
			return err
		}
		s.metrics.IncTCPConnections()
		go s.handleConn(ctx, conn)
	}
}

func (s *TCP) Shutdown(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener == nil {
		return nil
	}
	return s.listener.Close()
}

func (s *TCP) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		start := time.Now()
		payload, err := readFrame(reader)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				s.logger.Warn("tcp frame read failed", "remote", conn.RemoteAddr().String(), "error", err)
			}
			return
		}
		var req struct {
			Topic   string `json:"topic"`
			Payload string `json:"payload"`
			TraceID string `json:"trace_id"`
		}
		if err := json.Unmarshal(payload, &req); err != nil || req.Topic == "" {
			_ = writeFrame(conn, []byte(`{"status":"bad_request"}`))
			continue
		}
		traceCtx := observability.WithTraceID(ctx, req.TraceID)
		err = s.broker.Publish(traceCtx, broker.Message{Topic: req.Topic, Payload: req.Payload, TraceID: observability.TraceID(traceCtx)})
		if err != nil {
			_ = writeFrame(conn, []byte(`{"status":"unavailable"}`))
			continue
		}
		s.metrics.IncTCPFrames()
		s.metrics.ObserveLatency("tcp", time.Since(start))
		_ = writeFrame(conn, []byte(`{"status":"accepted"}`))
	}
}

func readFrame(r io.Reader) ([]byte, error) {
	var size uint32
	if err := binary.Read(r, binary.BigEndian, &size); err != nil {
		return nil, err
	}
	if size == 0 || size > 1<<20 {
		return nil, errors.New("invalid frame size")
	}
	payload := make([]byte, size)
	_, err := io.ReadFull(r, payload)
	return payload, err
}

func writeFrame(w io.Writer, payload []byte) error {
	if err := binary.Write(w, binary.BigEndian, uint32(len(payload))); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}
