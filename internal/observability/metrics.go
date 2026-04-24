package observability

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Metrics struct {
	started             time.Time
	httpRequests        atomic.Uint64
	queuePublished      atomic.Uint64
	queueProcessed      atomic.Uint64
	queueFailed         atomic.Uint64
	queueDeadLettered   atomic.Uint64
	tcpConnections      atomic.Uint64
	tcpFrames           atomic.Uint64
	udpDatagrams        atomic.Uint64
	udpMalformed        atomic.Uint64
	protocolLatencyLock sync.Mutex
	protocolLatency     map[string][]time.Duration
}

func NewMetrics() *Metrics {
	return &Metrics{started: time.Now(), protocolLatency: make(map[string][]time.Duration)}
}

func (m *Metrics) IncHTTPRequests()      { m.httpRequests.Add(1) }
func (m *Metrics) IncQueuePublished()    { m.queuePublished.Add(1) }
func (m *Metrics) IncQueueProcessed()    { m.queueProcessed.Add(1) }
func (m *Metrics) IncQueueFailed()       { m.queueFailed.Add(1) }
func (m *Metrics) IncQueueDeadLettered() { m.queueDeadLettered.Add(1) }
func (m *Metrics) IncTCPConnections()    { m.tcpConnections.Add(1) }
func (m *Metrics) IncTCPFrames()         { m.tcpFrames.Add(1) }
func (m *Metrics) IncUDPDatagrams()      { m.udpDatagrams.Add(1) }
func (m *Metrics) IncUDPMalformed()      { m.udpMalformed.Add(1) }

func (m *Metrics) ObserveLatency(protocol string, d time.Duration) {
	m.protocolLatencyLock.Lock()
	defer m.protocolLatencyLock.Unlock()
	m.protocolLatency[protocol] = append(m.protocolLatency[protocol], d)
}

func (m *Metrics) WritePrometheus(w io.Writer) {
	lines := []string{
		"# HELP netops_uptime_seconds Process uptime.",
		"# TYPE netops_uptime_seconds gauge",
		fmt.Sprintf("netops_uptime_seconds %.0f", time.Since(m.started).Seconds()),
		counter("netops_http_requests_total", "HTTP requests.", m.httpRequests.Load()),
		counter("netops_queue_published_total", "Messages accepted into the queue.", m.queuePublished.Load()),
		counter("netops_queue_processed_total", "Messages processed successfully.", m.queueProcessed.Load()),
		counter("netops_queue_failed_total", "Message processing failures.", m.queueFailed.Load()),
		counter("netops_queue_dead_lettered_total", "Messages moved to the dead letter queue.", m.queueDeadLettered.Load()),
		counter("netops_tcp_connections_total", "Accepted TCP connections.", m.tcpConnections.Load()),
		counter("netops_tcp_frames_total", "Length-prefixed TCP frames handled.", m.tcpFrames.Load()),
		counter("netops_udp_datagrams_total", "UDP datagrams handled.", m.udpDatagrams.Load()),
		counter("netops_udp_malformed_total", "Malformed UDP datagrams.", m.udpMalformed.Load()),
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			fmt.Fprintln(w, line)
			continue
		}
		fmt.Fprintln(w, line)
	}
	m.writeLatency(w)
}

func (m *Metrics) writeLatency(w io.Writer) {
	m.protocolLatencyLock.Lock()
	defer m.protocolLatencyLock.Unlock()
	protocols := make([]string, 0, len(m.protocolLatency))
	for protocol := range m.protocolLatency {
		protocols = append(protocols, protocol)
	}
	sort.Strings(protocols)
	fmt.Fprintln(w, "# HELP netops_protocol_latency_seconds Average protocol handling latency.")
	fmt.Fprintln(w, "# TYPE netops_protocol_latency_seconds gauge")
	for _, protocol := range protocols {
		var total time.Duration
		for _, d := range m.protocolLatency[protocol] {
			total += d
		}
		avg := float64(total) / float64(len(m.protocolLatency[protocol])) / float64(time.Second)
		fmt.Fprintf(w, "netops_protocol_latency_seconds{protocol=%q} %.6f\n", protocol, avg)
	}
}

func counter(name, help string, value uint64) string {
	return fmt.Sprintf("# HELP %s %s\n# TYPE %s counter\n%s %d", name, help, name, name, value)
}
