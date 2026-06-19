resource "e2b_sandbox" "example" {
  template_id = "base"
  timeout     = 300

  metadata = {
    managed_by = "terraform"
  }
}
