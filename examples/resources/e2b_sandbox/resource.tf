variable "egress_proxy_password" {
  description = "Optional SOCKS5 proxy password for sandbox egress."
  type        = string
  sensitive   = true
}

resource "e2b_sandbox" "example" {
  template_id = "base"
  timeout     = 300
  auto_resume = true

  allow_internet_access        = true
  network_allow_public_traffic = false
  network_allow_out            = ["api.example.com"]
  network_deny_out             = ["ALL_TRAFFIC"]
  network_egress_proxy = {
    address  = "proxy.example.com:1080"
    username = "sandbox-egress"
    password = var.egress_proxy_password
  }
  network_mask_request_host = "sandbox.example.com"
  network_rules = {
    "api.example.com" = [
      {
        headers = {
          "X-E2B-Policy" = "terraform-managed"
        }
      }
    ]
  }

  volume_mounts = [
    {
      name = "cache"
      path = "/mnt/cache"
    }
  ]

  metadata = {
    managed_by = "terraform"
  }
}
