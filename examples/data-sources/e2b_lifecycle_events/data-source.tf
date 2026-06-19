data "e2b_lifecycle_events" "recent" {
  sandbox_id = "sbx_123"
  types      = ["sandbox.started", "sandbox.paused"]
  limit      = 50
  order_asc  = false
}
