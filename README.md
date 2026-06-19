# Terraform Provider for E2B

Terraform provider for [E2B](https://e2b.dev), built with the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

This provider covers E2B Platform API objects that are useful for sandbox operations and team governance:

- `e2b_access_token` resource
- `e2b_api_key` resource
- `e2b_lifecycle_webhook` resource
- `e2b_sandbox` resource
- `e2b_snapshot` resource
- `e2b_template` resource
- `e2b_template_tags` resource
- `e2b_volume` resource
- `e2b_api_key`, `e2b_api_keys`, `e2b_lifecycle_events`, `e2b_lifecycle_webhook`, `e2b_lifecycle_webhooks`, `e2b_sandbox`, `e2b_sandboxes`, `e2b_snapshots`, `e2b_template`, `e2b_template_tags`, `e2b_templates`, `e2b_team_metric_max`, `e2b_team_metrics`, `e2b_teams`, `e2b_volume`, and `e2b_volumes` data sources

Every data source with a creatable E2B API object has a matching Terraform resource. The exceptions are read-only public API surfaces: `e2b_teams`, `e2b_team_metrics`, `e2b_team_metric_max`, and `e2b_lifecycle_events`.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24
- E2B API key exported as `E2B_API_KEY` for standard sandbox, template, volume, lifecycle, snapshot, and tag routes
- E2B access token exported as `E2B_ACCESS_TOKEN` for bearer-authenticated team, access token, and API key routes
- E2B team ID exported as `E2B_TEAM_ID` for team-scoped bearer routes such as API key management

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
- `E2B_ACCESS_TOKEN` for bearer-authenticated team routes.
- `E2B_TEAM_ID` for team-scoped bearer-authenticated routes.
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
