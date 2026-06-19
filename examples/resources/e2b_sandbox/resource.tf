resource "e2b_sandbox" "example" {
  template_id = "base"
  timeout     = 300
  auto_resume = true

  network_allow_public_traffic = false
  network_allow_out            = ["8.8.8.8/32"]
  network_deny_out             = ["203.0.113.0/24"]
  network_mask_request_host    = "sandbox.example.com"

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
