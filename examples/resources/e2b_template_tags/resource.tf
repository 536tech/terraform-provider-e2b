resource "e2b_template_tags" "production" {
  template_id = "tpl_123"
  target      = "example-template:build-123"
  name        = "example-template"
  tags        = ["production", "stable"]
}
