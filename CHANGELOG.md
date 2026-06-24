## Unreleased

FEATURES:

- Added sandbox metrics, sandbox metric history, sandbox logs, and template alias data sources from the current E2B OpenAPI spec.

BREAKING CHANGES:

- Removed `E2B_ACCESS_TOKEN` provider authentication and the legacy access-token/API-key management resources because E2B deprecated access tokens and those management routes are no longer in the public OpenAPI spec.

## 0.1.0 (2026-06-19)

FEATURES:

- Initial E2B provider using Terraform Plugin Framework.
- Added `e2b_access_token`, `e2b_api_key`, `e2b_lifecycle_webhook`, `e2b_sandbox`, `e2b_snapshot`, `e2b_template`, `e2b_template_tags`, and `e2b_volume` resources.
- Added data sources for API keys, lifecycle events and webhooks, sandboxes, snapshots, templates, template tags, teams, team metrics, and volumes.
- Added API key, access token, and team-scoped bearer provider authentication.
- Added generated provider documentation and examples.
- Added unit and acceptance test coverage for supported E2B resource and data source surfaces.
