data "e2b_sandbox_logs" "recent" {
  sandbox_id = "sbx_123"
  limit      = 100
  direction  = "backward"
}
