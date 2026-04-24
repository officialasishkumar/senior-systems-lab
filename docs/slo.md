# SLO and Alerting

## Service Level Objectives

| Signal | Objective | Window |
| --- | --- | --- |
| Availability | 99.9 percent successful health-checked service availability | 30 days |
| HTTP latency | 95 percent of publish requests under 150 ms | 7 days |
| TCP latency | 95 percent of accepted frames under 100 ms | 7 days |
| Queue durability | 99.99 percent of accepted messages processed or dead-lettered | 30 days |

## Error Budget Policy

- Page when the 2-hour burn rate indicates the 30-day budget will be exhausted.
- Freeze risky deploys when more than 50 percent of the monthly budget is consumed.
- Require a mitigation plan for repeat incidents in the same failure domain.

## Alerts

- High 5xx or unavailable responses.
- Queue publish failures above 1 percent for 5 minutes.
- Dead-letter growth above baseline.
- TCP connection surge with rising frame errors.
- UDP malformed datagrams above baseline.
- Pod restart loop or readiness failures.

## Dashboards

- RED: request rate, error rate, duration.
- USE: CPU, memory, network, goroutines, file descriptors.
- Queue: published, processed, failed, dead-lettered, current depth.
- Protocols: HTTP requests, TCP frames, UDP datagrams, malformed UDP.

