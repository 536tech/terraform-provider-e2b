variable "address_space" {
  description = "CIDR ranges for the E2B Azure readiness virtual network."
  type        = list(string)
  default     = ["10.42.0.0/16"]
}

variable "environment" {
  description = "Deployment environment name."
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}

variable "location" {
  description = "Azure region for the readiness resources."
  type        = string
  default     = "eastus"
}

variable "project_name" {
  description = "Short name used in Azure resource names."
  type        = string
  default     = "e2b"

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{1,11}$", var.project_name))
    error_message = "Project name must be 2-12 characters, start with a lowercase letter, and contain only lowercase letters, numbers, or hyphens."
  }
}

variable "subnet_prefixes" {
  description = "Subnet CIDR ranges for edge, orchestrator, build, and data workloads."
  type        = map(string)
  default = {
    build        = "10.42.3.0/24"
    data         = "10.42.4.0/24"
    edge         = "10.42.1.0/24"
    orchestrator = "10.42.2.0/24"
  }
}
