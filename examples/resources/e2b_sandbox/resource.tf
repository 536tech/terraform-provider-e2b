resource "e2b_sandbox" "example" {
  template_id = "base"
  timeout     = 300

  network_allow_out = ["8.8.8.8/32"]
  network_deny_out  = ["203.0.113.0/24"]

  metadata = {
    managed_by = "terraform"
  }
}
