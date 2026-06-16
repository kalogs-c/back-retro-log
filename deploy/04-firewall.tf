resource "google_compute_firewall" "backretrolog-http" {
  name    = "${var.vm_name}-http"
  project = var.project_id
  network = var.network

  allow {
    protocol = "tcp"
    ports    = ["80"]
  }

  source_ranges = var.allowed_source_ranges
  target_tags   = ["backretrolog-vm"]
}
