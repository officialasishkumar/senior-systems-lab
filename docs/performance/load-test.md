# Performance Test Plan

## HTTP

Use a steady publish workload and watch p95 latency, queue depth, and publish failures:

```bash
vegeta attack -duration=60s -rate=500/s -header "Content-Type: application/json" \
  -body <(printf '{"topic":"load.http","payload":"ok"}') \
  -method POST http://127.0.0.1:8080/queue/publish | vegeta report
```

## TCP

Run many concurrent `netprobe` clients against the framed TCP endpoint and compare accepted frames with worker throughput:

```bash
seq 1 1000 | xargs -P 50 -I{} go run ./cmd/netprobe -mode tcp -addr 127.0.0.1:9090 -topic load.tcp -payload {}
```

## UDP

UDP heartbeats should be measured by packet loss and malformed packet rate, not only latency:

```bash
seq 1 10000 | xargs -P 100 -I{} go run ./cmd/netprobe -mode udp -addr 127.0.0.1:9091 >/dev/null
```

## Benchmarks

```bash
go test -bench=. ./...
go test -race ./...
```

