# Threat Model

## Assets

- Message payloads.
- Trace and log metadata.
- Deployment credentials.
- Container images and release artifacts.
- Runtime configuration.

## Controls

- Non-root distroless runtime image.
- Read-only Kubernetes filesystem.
- Dropped Linux capabilities.
- Service account token automount disabled.
- Bounded request and frame sizes.
- Input validation on HTTP, TCP, and UDP paths.
- Vulnerability scanning and SBOM generation in CI.
- Secrets kept outside the repository.

## Risks

- Oversized TCP frames can exhaust memory if not bounded.
- UDP can be spoofed without network-level source controls.
- Queue saturation can amplify caller retries.
- Logs can leak sensitive payloads if payload logging is added without redaction.
- Supply-chain risk remains if image and dependency scanning are not enforced.

## Mitigations

- Preserve the 1 MiB TCP frame limit.
- Restrict UDP ingress to private CIDR ranges.
- Rate limit public HTTP routes at the edge.
- Redact payload fields in log processors.
- Sign release images and verify signatures in deployment policy.

