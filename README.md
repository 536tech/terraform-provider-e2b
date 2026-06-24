# Terraform Provider for E2B

Terraform provider for [E2B](https://e2b.dev), built with the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

This provider covers E2B Platform API objects that are useful for sandbox operations and team governance:

- `e2b_lifecycle_webhook` resource
- `e2b_sandbox` resource
- `e2b_snapshot` resource
- `e2b_template` resource
- `e2b_template_tags` resource
- `e2b_volume` resource
- `e2b_lifecycle_events`, `e2b_lifecycle_webhook`, `e2b_lifecycle_webhooks`, `e2b_sandbox`, `e2b_sandbox_logs`, `e2b_sandbox_metric_history`, `e2b_sandbox_metrics`, `e2b_sandboxes`, `e2b_snapshots`, `e2b_template`, `e2b_template_alias`, `e2b_template_tags`, `e2b_templates`, `e2b_team_metric_max`, `e2b_team_metrics`, `e2b_volume`, and `e2b_volumes` data sources

Every data source with a creatable E2B API object has a matching Terraform resource. The exceptions are read-only public API surfaces: lifecycle events, sandbox logs/metrics, template aliases, and team metrics.

`E2B_ACCESS_TOKEN` is not supported. E2B announced access token deprecation on June 22, 2026; new token generation stops on July 1, 2026, and access-token-authenticated requests stop working on August 1, 2026. Use `E2B_API_KEY`.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24
- E2B API key exported as `E2B_API_KEY` for standard sandbox, template, volume, lifecycle, snapshot, and tag routes

## Building the Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Using the Provider

```hcl
terraform {
  required_providers {
    e2b = {
      source = "536tech/e2b"
    }
  }
}

provider "e2b" {}

resource "e2b_sandbox" "example" {
  template_id = "base"
  timeout     = 300
  auto_resume = true

  network_allow_public_traffic = false
  network_allow_out            = ["8.8.8.8/32"]
  network_deny_out             = ["203.0.113.0/24"]
  network_mask_request_host    = "sandbox.example.com"

  metadata = {
    managed_by = "terraform"
  }
}

resource "e2b_template" "example" {
  name       = "example-template"
  from_image = "e2bdev/base:latest"
  start_cmd  = "sh -c \"sleep 3600\""
  ready_cmd  = "true"
  cpu_count  = 2
  memory_mb  = 512
}
```

The provider reads:

- `E2B_API_KEY` for API authentication.
- `E2B_API_URL` to override the API base URL. The default is `https://api.e2b.app`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

Run unit tests:

```shell
go test ./...
```

Run live acceptance tests:

```shell
export E2B_API_KEY=...
make testacc
```

Acceptance tests create real E2B resources and clean them up. Volume acceptance tests are opt-in because E2B may not enable volumes on every account:

```shell
E2B_ACC_VOLUMES=1 make testacc
```

Generate documentation:

```shell
make generate
```

See [docs/api-coverage.md](docs/api-coverage.md) for the current OpenAPI coverage map.

The [Azure self-hosting readiness example](examples/use-cases/azure-self-hosting-readiness) prepares Azure networking, storage, registry, identity, secrets, and logging resources for future E2B Azure BYOC/self-hosting work. It is a readiness example, not a claim that Azure BYOC is generally available.
