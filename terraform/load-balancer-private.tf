resource "google_compute_global_forwarding_rule" "internal" {
  name = "${local.prefix}-forwarding-rule-internal"

  depends_on            = [google_compute_subnetwork.proxy_subnet_global]
  ip_protocol           = "TCP"
  load_balancing_scheme = "INTERNAL_MANAGED"
  port_range            = "80"
  target                = google_compute_target_http_proxy.internal.id

  subnetwork = google_compute_subnetwork.lb_subnet.id
}

resource "google_compute_backend_service" "internal" {
  name                            = "${local.prefix}-backend-service-internal"
  enable_cdn                      = false
  timeout_sec                     = 10
  connection_draining_timeout_sec = 10
  load_balancing_scheme           = "INTERNAL_MANAGED"

  backend {
    group          = google_compute_network_endpoint_group.default.id
    balancing_mode = "RATE"
    max_rate       = 1000
  }

  health_checks = [google_compute_health_check.default.id]
}

resource "google_compute_target_http_proxy" "internal" {
  name    = "${local.prefix}-http-proxy-internal"
  url_map = google_compute_url_map.internal.id
}

# URL map
resource "google_compute_url_map" "internal" {
  name            = "${local.prefix}-url-map-internal"
  default_service = google_compute_backend_service.internal.id
}