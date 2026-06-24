locals {
  name_prefix = replace("${var.project_name}-${var.environment}", "-", "")

  tags = {
    Environment = var.environment
    ManagedBy   = "Terraform"
    Project     = var.project_name
  }
}
