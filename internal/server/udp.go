package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/officialasishkumar/pulsemesh/internal/observability"
)

type UDP struct {
	addr    string
	metrics *observability.Metrics
	logger  *slog.Logger
	conn    *net.UDPConn
	mu      sync.Mutex
}

func NewUDP(addr string, metrics *observability.Metrics, logger *slog.Logger) *UDP {
	return &UDP{addr: addr, metrics: metrics, logger: logger}
}

func (s *UDP) Run(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.conn = conn
	s.mu.Unlock()
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()
	buf := make([]byte, 2048)
	for {
		start := time.Now()
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) || ctx.Err() != nil {
				return context.Canceled
			}
			return err
		}
		s.metrics.IncUDPDatagrams()
		reply := s.handleDatagram(strings.TrimSpace(string(buf[:n])))
		if _, err := conn.WriteToUDP([]byte(reply), remote); err != nil {
			s.logger.Warn("udp write failed", "remote", remote.String(), "error", err)
		}
		s.metrics.ObserveLatency("udp", time.Since(start))
	}
}

func (s *UDP) Shutdown(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == nil {
		return nil
	}
	return s.conn.Close()
}

func (s *UDP) handleDatagram(payload string) string {
	parts := strings.Fields(payload)
	if len(parts) == 0 {
		s.metrics.IncUDPMalformed()
		return "ERR empty datagram"
	}
	switch parts[0] {
	case "PING":
		return "PONG"
	case "HEARTBEAT":
		if len(parts) < 2 {
			s.metrics.IncUDPMalformed()
			return "ERR missing node id"
		}
		return "ACK " + parts[1]
	default:
		s.metrics.IncUDPMalformed()
		return "ERR unsupported command"
	}
}
