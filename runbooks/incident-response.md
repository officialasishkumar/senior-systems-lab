# Incident Response Runbook

## Triage

1. Check `/healthz`, `/readyz`, and `/metrics`.
2. Compare request rate, error rate, queue failures, and dead-letter growth against baseline.
3. Inspect recent deploys and image tags.
4. Check pod restarts, readiness events, CPU, memory, and network saturation.
5. Identify whether the issue is ingress, queue, worker, downstream, or deployment related.

## Queue Saturation

1. Confirm `netops_queue_published_total` is still increasing.
2. Check publish failures and queue depth.
3. Scale replicas or workers if CPU and downstream capacity allow it.
4. Shed low-priority traffic at the edge if the queue remains full.
5. Inspect dead letters and replay only idempotent messages.

## TCP Ingestion Failures

1. Check L4 load balancer health and connection count.
2. Confirm clients send a 4-byte big-endian frame length followed by JSON.
3. Look for invalid frame-size warnings.
4. Drain connections before rolling restarts.

## UDP Heartbeat Failures

1. Verify UDP listener exposure and security group rules.
2. Confirm packet path with private-source clients.
3. Compare malformed datagrams against baseline.
4. Treat missing heartbeats as suspect until multiple intervals are missed.

## Rollback

1. Stop active rollout.
2. Deploy previous known-good image tag.
3. Watch readiness, error rate, queue depth, and dead letters for 15 minutes.
4. Capture logs, metrics, timeline, and contributing factors.

