resource "google_compute_instance" "default" {
  name         = "${local.prefix}-vm"
  machine_type = "f1-micro"
  zone         = "${var.region}-a"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
    }
  }

  network_interface {
    network    = google_compute_network.default.id
    subnetwork = google_compute_subnetwork.backend_subnet.id

    access_config {
      // Ephemeral public IP
    }
  }

  deletion_protection = false
}