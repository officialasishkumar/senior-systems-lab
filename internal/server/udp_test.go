package server

import (
	"log/slog"
	"testing"

	"github.com/officialasishkumar/senior-systems-lab/internal/observability"
)

func TestUDPCommands(t *testing.T) {
	s := NewUDP(":0", observability.NewMetrics(), slog.Default())
	tests := map[string]string{
		"PING":             "PONG",
		"HEARTBEAT node-1": "ACK node-1",
		"HEARTBEAT":        "ERR missing node id",
		"UNKNOWN":          "ERR unsupported command",
	}
	for input, want := range tests {
		if got := s.handleDatagram(input); got != want {
			t.Fatalf("handleDatagram(%q) = %q, want %q", input, got, want)
		}
	}
}
