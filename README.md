# PulseMesh

PulseMesh is a Go event ingress and node heartbeat platform for operating distributed systems. It accepts application events over HTTP and TCP, accepts lightweight node heartbeats over UDP, routes accepted work through a bounded broker, and exposes the operational signals needed to run it safely.

- HTTP APIs for health, readiness, queue publishing, dead-letter inspection, and Prometheus metrics.
- A bounded in-memory message broker with worker pools, retries, trace propagation, and dead-letter handling.
- A length-prefixed TCP protocol for framed message ingestion.
- A UDP heartbeat protocol for low-overhead liveness telemetry.
- Structured JSON logs, correlation IDs, graceful shutdown, and operational runbooks.
- Docker, Compose, Kubernetes, Helm, Terraform, CI, release automation, SBOM generation, and image scanning.

## Architecture

```text
clients
  |-- HTTP :8080 /queue/publish /metrics /healthz /readyz
  |-- TCP  :9090 length-prefixed JSON frames
  |-- UDP  :9091 PING and HEARTBEAT datagrams
        |
        v
 bounded broker -> workers -> retry policy -> dead-letter queue
        |
        v
 structured logs + Prometheus metrics + trace IDs
```

The code is intentionally dependency-light so a fresh checkout builds quickly and the core systems behavior stays visible in the source. Production integrations such as Kafka, NATS, Redis Streams, OpenTelemetry collectors, and managed databases can replace the local interfaces without changing the service boundaries.

## Run Locally

```bash
make test
make e2e
make build
./bin/pulsemesh
```

In another terminal:

```bash
make probe-http
make probe-tcp
make probe-udp
curl -fsS http://127.0.0.1:8080/metrics
```

## Docker

```bash
make docker-build
docker run --rm -p 8080:8080 -p 9090:9090 -p 9091:9091/udp pulsemesh:local
```

## Deploy

```bash
kubectl apply -k deploy/kubernetes
helm lint deploy/helm/pulsemesh
helm upgrade --install pulsemesh deploy/helm/pulsemesh
terraform -chdir=infra/terraform init
terraform -chdir=infra/terraform validate
```

## Capabilities

- HTTP event publishing with validation, request timeouts, trace propagation, and consistent JSON responses.
- Bounded message ingestion with worker pools, backpressure, retries, dead-letter storage, queue depth, and failure isolation.
- TCP event ingestion with explicit framing, binary length prefixes, payload validation, connection lifecycle handling, and fuzz coverage.
- UDP heartbeat handling for stateless liveness telemetry, malformed packet accounting, and packet-level tests.
- Prometheus metrics, structured logs, trace IDs, protocol latency signals, and scrape-ready deployment configuration.
- Health and readiness endpoints, SLOs, error budget policy, runbooks, rollback checklist, and incident templates.
- Docker, Compose, Kubernetes, Helm, Terraform, GitHub Actions, SBOM generation, vulnerability scanning, and release automation.
- VPC/subnet layout, route tables, security groups, TCP/UDP ingress rules, NAT/load balancer/DNS notes, and TLS/mTLS design.
- Least-privilege service account, non-root container, read-only filesystem, input validation, secrets policy, threat model, and scanner-backed checks.
- Unit tests, race tests, fuzz target, E2E protocol tests, CI validation, deployment manifest rendering, OpenAPI, load-test plan, and chaos-test matrix.

## API Examples

```bash
curl -fsS -X POST http://127.0.0.1:8080/queue/publish \
  -H 'Content-Type: application/json' \
  -H 'X-Trace-ID: demo-trace' \
  -d '{"topic":"orders.created","payload":"order-123"}'
```

TCP frames are encoded as four big-endian bytes containing the JSON payload size followed by the JSON body:

```json
{"topic":"orders.created","payload":"order-123","trace_id":"demo-trace"}
```

UDP commands:

```text
PING
HEARTBEAT node-1
```
