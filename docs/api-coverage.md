# E2B OpenAPI coverage

Checked against `https://raw.githubusercontent.com/e2b-dev/docs/main/openapi-public.yml` on 2026-06-24.

- Operations: 65
- SHA-256: `1eb15d10fda03782929b258d9ae870916f34b5bfaa3cb63e79d54ad1d8693260`

## Covered platform API surfaces

| Surface | Terraform coverage |
| --- | --- |
| Sandboxes | `e2b_sandbox` resource; `e2b_sandbox`, `e2b_sandboxes` data sources |
| Sandbox network and timeout updates | `e2b_sandbox` update |
| Sandbox metrics and logs | `e2b_sandbox_metrics`, `e2b_sandbox_metric_history`, `e2b_sandbox_logs` |
| Templates | `e2b_template` resource; `e2b_template`, `e2b_templates`, `e2b_template_alias` data sources |
| Template builds | `e2b_template` resource |
| Template tags | `e2b_template_tags` resource and data source |
| Volumes | `e2b_volume` resource; `e2b_volume`, `e2b_volumes` data sources |
| Snapshots | `e2b_snapshot` resource; `e2b_snapshots` data source |
| Lifecycle webhooks and events | `e2b_lifecycle_webhook` resource; `e2b_lifecycle_webhook`, `e2b_lifecycle_webhooks`, `e2b_lifecycle_events` data sources |
| Team metrics | `e2b_team_metrics`, `e2b_team_metric_max` |

`GET /templates/{templateID}/builds/{buildID}/status` is marked `AccessTokenAuth` in OpenAPI, but the live API accepts `E2B_API_KEY`; `TestAccTemplateResource` covers that API-key path.

## Intentionally excluded

| OpenAPI surface | Reason |
| --- | --- |
| `GET /health`, env vars, files, filesystem RPC, and process RPC routes | Runtime data-plane calls against a running sandbox. Use the E2B SDK or CLI rather than Terraform state. |
| Sandbox `pause`, deprecated `resume`, `connect`, and `refreshes` operations | Operational lifecycle actions, not durable desired state. Terraform creates, reads, updates supported configuration, and deletes sandboxes. |
| Deprecated template v1/v2 create/rebuild/update routes | The provider uses non-deprecated template creation and v2 build start routes. |
| `GET /templates/{templateID}/files/{hash}` | Build upload-link helper, not a durable Terraform object. |
| `GET /templates/{templateID}/builds/{buildID}/logs` | Listed in OpenAPI, but live API returned `404 validation error: method not allowed` during acceptance with API-key auth. |
| `GET /teams` | Requires `AccessTokenAuth`; E2B access-token authentication is deprecated and stops working on 2026-08-01. |

## Removed legacy surfaces

The current public OpenAPI spec no longer includes `/api-keys` or `/access-tokens`. Live probes with `E2B_API_KEY` returned:

- `GET /api-keys`: `401`
- `GET /access-tokens`: `404`

The provider therefore removed `E2B_ACCESS_TOKEN` configuration plus the legacy `e2b_access_token`, `e2b_api_key`, and `e2b_api_keys` Terraform surfaces.
