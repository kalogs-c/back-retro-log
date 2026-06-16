variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "region" {
  type        = string
  description = "GCP region (us-central1, us-west1, us-east1 for free tier)"
  default     = "us-central1"
}

variable "zone" {
  type        = string
  description = "GCP zone"
  default     = "us-central1-a"
}

variable "vm_name" {
  type        = string
  description = "VM instance name"
}

variable "ip_name" {
  type        = string
  description = "Static external IP name"
}

variable "machine_type" {
  type        = string
  description = "GCE machine type (e2-micro is free tier eligible)"
  default     = "e2-micro"
}

variable "network" {
  type        = string
  description = "VPC network name"
  default     = "default"
}

variable "disk_size" {
  type        = number
  description = "Persistent disk size for SQLite data in GB (free tier: 30 GB total with boot disk)"
  default     = 10

  validation {
    condition     = var.disk_size >= 5 && var.disk_size <= 100
    error_message = "Disk size must be between 5 and 100 GB."
  }
}

variable "docker_image" {
  type        = string
  description = "Docker image to deploy (e.g. kalogsc/back-retro-log:v1.0.0)"
}

variable "rawg_api_key" {
  type        = string
  description = "RAWG API key (optional — dummy provider used if empty)"
  default     = ""
  sensitive   = true
}

variable "base_url" {
  type        = string
  description = "Public base URL for invite links (auto-detected from external IP if not set)"
  default     = ""
}

variable "allowed_source_ranges" {
  type        = list(string)
  description = "CIDR ranges allowed to access the app (default: 0.0.0.0/0 = anywhere)"
  default     = ["0.0.0.0/0"]
}
