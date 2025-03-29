resource "google_compute_router" "default" {
  name    = "${local.prefix}-router"
  region  = google_compute_subnetwork.backend_subnet.region
  network = google_compute_network.default.id
}

resource "google_compute_router_nat" "nat" {
  name                               = "${local.prefix}-router-nat"
  router                             = google_compute_router.default.name
  region                             = google_compute_router.default.region
  nat_ip_allocate_option             = "MANUAL_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  nat_ips = [google_compute_address.nat.self_link]
}

resource "google_compute_address" "nat" {
  name         = "${local.prefix}-nat"
  address_type = "EXTERNAL"
  region       = google_compute_router.default.region
}
