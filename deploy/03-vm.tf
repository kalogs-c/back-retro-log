data "google_compute_image" "cos" {
  family  = "cos-stable"
  project = "cos-cloud"
}

resource "google_compute_disk" "backretrolog_data" {
  name    = "${var.vm_name}-data"
  project = var.project_id
  type    = "pd-standard"
  zone    = var.zone
  size    = var.disk_size
}

resource "google_compute_instance" "backretrolog-vm" {
  name                      = var.vm_name
  machine_type              = var.machine_type
  zone                      = var.zone
  allow_stopping_for_update = true

  boot_disk {
    initialize_params {
      image = data.google_compute_image.cos.self_link
      size  = 10
      type  = "pd-standard"
    }
  }

  attached_disk {
    source      = google_compute_disk.backretrolog_data.id
    device_name = "backretrolog-data"
  }

  metadata_startup_script = templatefile("${path.module}/startup.sh.tftpl", {
    docker_image = var.docker_image
    base_url     = var.base_url
    rawg_api_key = var.rawg_api_key
  })

  labels = {
    app = "backretrolog"
  }

  tags = ["backretrolog-vm"]

  network_interface {
    network = var.network
    access_config {
      nat_ip = google_compute_address.backretrolog.address
    }
  }

  service_account {
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }
}
