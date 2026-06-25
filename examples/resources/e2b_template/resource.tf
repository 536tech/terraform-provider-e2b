resource "e2b_template" "example" {
  name       = "example-template"
  from_image = "e2bdev/base:latest"
  start_cmd  = "sh -c \"sleep 3600\""
  ready_cmd  = "true"
  cpu_count  = 2
  memory_mb  = 512
  public     = false
}
