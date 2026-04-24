# Network Design

## Topology

- Public subnets host internet-facing load balancers.
- Private subnets host application pods or compute nodes.
- NAT gateways provide controlled outbound access for private workloads.
- Route tables keep east-west traffic private and north-south ingress explicit.
- Security groups allow only required HTTP, TCP, and UDP ports from trusted CIDR ranges.

## Load Balancing

- HTTP uses L7 load balancing for path-aware routing, TLS termination, request logs, and health checks.
- TCP uses L4 load balancing to preserve stream behavior and reduce protocol interference.
- UDP uses L4 load balancing with short idle timeouts and idempotent heartbeat semantics.
- Connection draining is required for rolling deployments so TCP clients are not cut during deploys.

## DNS and TLS

- Public DNS points to the HTTP load balancer.
- Internal DNS names expose TCP and UDP services to private clients.
- TLS terminates at the edge for HTTP.
- mTLS is appropriate for service-to-service TCP ingestion when clients are known workloads.

## Firewall Rules

| Port | Protocol | Purpose | Source |
| --- | --- | --- | --- |
| 8080 | TCP | HTTP API and metrics | load balancer or private CIDR |
| 9090 | TCP | framed ingestion | private CIDR |
| 9091 | UDP | heartbeat telemetry | private CIDR |

