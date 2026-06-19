# Terraform Provider for E2B

Terraform provider for [E2B](https://e2b.dev), built with the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

This provider currently covers the E2B Platform API objects that can be tested with an `E2B_API_KEY`:

- `e2b_sandbox` resource
- `e2b_volume` resource
- `e2b_sandbox`, `e2b_sandboxes`, `e2b_template`, `e2b_templates`, `e2b_volume`, and `e2b_volumes` data sources

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24
- E2B API key exported as `E2B_API_KEY`

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

  network_allow_out = ["8.8.8.8/32"]
  network_deny_out  = ["203.0.113.0/24"]

  metadata = {
    managed_by = "terraform"
  }
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
