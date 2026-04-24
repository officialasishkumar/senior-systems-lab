#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

HTTP_ADDR="127.0.0.1:18080" TCP_ADDR="127.0.0.1:19090" UDP_ADDR="127.0.0.1:19091" go run ./cmd/pulsemesh > /tmp/pulsemesh-e2e.log 2>&1 &
pid=$!
trap 'kill "$pid" >/dev/null 2>&1 || true' EXIT

for _ in $(seq 1 30); do
  if curl -fsS http://127.0.0.1:18080/readyz >/dev/null 2>&1; then
    break
  fi
  sleep 0.2
done

curl -fsS http://127.0.0.1:18080/healthz >/dev/null
go run ./cmd/netprobe -mode http -addr 127.0.0.1:18080 -topic e2e.http -payload ok >/dev/null
go run ./cmd/netprobe -mode tcp -addr 127.0.0.1:19090 -topic e2e.tcp -payload ok | grep -q accepted
go run ./cmd/netprobe -mode udp -addr 127.0.0.1:19091 | grep -q PONG
curl -fsS http://127.0.0.1:18080/metrics | grep -q netops_http_requests_total
curl -fsS http://127.0.0.1:18080/queue/stats | grep -q queued
