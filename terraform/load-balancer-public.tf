# This file creates a forwarding rule with a public IP

resource "google_compute_global_forwarding_rule" "external" {
  name = "${local.prefix}-forwarding-rule-external"

  depends_on            = [google_compute_subnetwork.proxy_subnet]
  ip_protocol           = "TCP"
  load_balancing_scheme = "EXTERNAL_MANAGED"
  port_range            = "80"
  target                = google_compute_target_http_proxy.external.id
}

resource "google_compute_backend_service" "external" {
  name                            = "${local.prefix}-backend-service"
  enable_cdn                      = false
  timeout_sec                     = 10
  connection_draining_timeout_sec = 10
  load_balancing_scheme           = "EXTERNAL_MANAGED"

  backend {
    group          = google_compute_network_endpoint_group.default.id
    balancing_mode = "RATE"
    max_rate       = 1000
  }

  health_checks = [google_compute_health_check.default.id]
}

resource "google_compute_target_http_proxy" "external" {
  name    = "${local.prefix}-http-proxy"
  url_map = google_compute_url_map.external.id
}

# URL map
resource "google_compute_url_map" "external" {
  name            = "${local.prefix}-url-map"
  default_service = google_compute_backend_service.external.id
}

