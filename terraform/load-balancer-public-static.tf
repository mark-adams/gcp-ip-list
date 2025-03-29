resource "google_compute_global_forwarding_rule" "external-static-ip" {
  name = "${local.prefix}-forwarding-rule-external-static"

  depends_on            = [google_compute_subnetwork.proxy_subnet]
  ip_protocol           = "TCP"
  load_balancing_scheme = "EXTERNAL_MANAGED"
  port_range            = "80"
  target                = google_compute_target_http_proxy.external.id

  ip_address = google_compute_global_address.default.id
}

resource "google_compute_global_address" "default" {
  name = "${local.prefix}-static-address"
}