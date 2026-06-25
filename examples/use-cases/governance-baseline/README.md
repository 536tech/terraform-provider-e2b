# E2B governance baseline

This example composes public E2B API resources into a small governance baseline:

- Keep approved templates private unless explicitly published.
- Create sandboxes with no public inbound traffic.
- Default-deny outbound traffic except approved domains.
- Route sandbox egress through a SOCKS5 proxy.
- Attach per-domain request header transforms.
- Deliver lifecycle events to an audit webhook.
- Read sandbox and team metric evidence for compliance checks.

It is intentionally provider-only. It does not create the proxy, SIEM endpoint, or policy engine that receives the audit data.

