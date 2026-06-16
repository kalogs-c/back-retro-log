resource "google_compute_address" "backretrolog" {
  name    = var.ip_name
  project = var.project_id
  region  = var.region
}
