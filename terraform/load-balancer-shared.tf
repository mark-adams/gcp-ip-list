
resource "google_compute_network_endpoint_group" "default" {
  name         = "${local.prefix}-neg"
  network      = google_compute_network.default.id
  subnetwork   = google_compute_subnetwork.lb_subnet.id
  default_port = "90"
  zone         = "${var.region}-a"
}

resource "google_compute_health_check" "default" {
  name = "${local.prefix}-tcp-health-check"

  timeout_sec        = 1
  check_interval_sec = 1

  tcp_health_check {
    port = "80"
  }
}