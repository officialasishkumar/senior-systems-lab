# System Design

## Goals

PulseMesh is an event ingress and node heartbeat platform for distributed systems. Its core job is to accept events over HTTP and TCP, accept low-cost liveness heartbeats over UDP, apply backpressure when the system is saturated, and expose clear signals for operators.

## Request Flow

1. HTTP, TCP, or UDP traffic enters the process.
2. HTTP and TCP requests receive a trace ID from the caller or generate one at ingress.
3. Message-producing paths publish to a bounded broker.
4. Worker goroutines consume messages with retry and dead-letter behavior.
5. Metrics and logs record protocol, queue, and worker outcomes.

## Queue Semantics

- Delivery model: at-least-once inside the process.
- Ordering: FIFO at queue entry, with concurrent workers allowing completion reordering.
- Backpressure: bounded channel capacity returns `503` or protocol-level unavailable replies when full.
- Retries: three attempts with short increasing delay.
- Dead letters: failed messages are retained for inspection.

External production equivalents include Kafka for durable ordered partitions, NATS JetStream for lightweight work queues, RabbitMQ for routing-heavy workflows, or Redis Streams for small operational pipelines.

## TCP Design

TCP is stream-oriented, so the service uses a length-prefixed frame:

- 4-byte unsigned big-endian payload length.
- JSON payload body.
- 1 MiB frame size limit to prevent unbounded allocation.
- Per-connection goroutine for isolation.

This keeps framing, partial-read handling, malformed frames, backpressure, and connection lifecycle management explicit.

## UDP Design

UDP is datagram-oriented and lossy. This service uses it only for idempotent liveness telemetry:

- `PING` returns `PONG`.
- `HEARTBEAT node-id` returns `ACK node-id`.
- Malformed commands increment a metric and return a small error.

UDP is not used for critical state transitions because packets can be dropped, duplicated, or reordered.

## Scaling Strategy

- Scale horizontally behind an L7 load balancer for HTTP and an L4 load balancer for TCP/UDP.
- Keep queues bounded per pod to avoid memory exhaustion.
- Increase worker count until CPU or downstream saturation appears.
- Use queue depth, publish failure rate, worker latency, and TCP connection count as autoscaling inputs.
- Move broker state to Kafka/NATS when messages must survive pod restarts.

## Failure Modes

- Queue saturation: publish returns unavailable and metrics expose queue pressure.
- Slow consumers: queue depth rises; workers can be scaled until downstream limits are reached.
- TCP clients disconnect mid-frame: read fails and the connection is closed.
- UDP packet loss: caller retries heartbeat on its own cadence.
- Pod termination: graceful shutdown drains listeners and stops accepting new work.
- Observability outage: service continues serving traffic while local logs and metrics remain available.

## Capacity Model

Core inputs:

- Requests per second by protocol.
- Average payload size.
- Worker processing latency.
- Retry rate.
- Queue capacity.

Sizing formula:

```text
required_workers = ceil(target_rps * average_processing_seconds / target_utilization)
queue_seconds    = queue_capacity / publish_rate
```

For example, 1,000 messages per second at 20 ms processing time and 70 percent worker utilization requires about 29 workers.
