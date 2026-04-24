# Failure Test Matrix

| Scenario | Expected Behavior | Signal |
| --- | --- | --- |
| Queue at capacity | HTTP returns 503, TCP returns unavailable | publish failures, queue depth |
| Worker handler failures | message retries three times, then moves to dead-letter queue | failed and dead-letter counters |
| TCP partial frame | connection closes without process crash | warning log, no panic |
| Oversized TCP frame | frame rejected before allocation beyond limit | warning log |
| UDP malformed command | error reply and malformed counter increment | malformed UDP metric |
| Pod termination | listeners close and process exits cleanly | readiness drops, no restart loop |
| Rolling deploy | maxUnavailable zero preserves capacity | deployment events |
| Network ACL blocks UDP | heartbeat loss is isolated from HTTP/TCP | UDP datagram drop |

