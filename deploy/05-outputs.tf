output "instance_name" {
  value       = google_compute_instance.backretrolog-vm.name
  description = "VM instance name"
}

output "external_ip" {
  value       = google_compute_address.backretrolog.address
  description = "External IP address"
}

output "app_url" {
  value       = "http://${google_compute_address.backretrolog.address}"
  description = "App URL"
}
