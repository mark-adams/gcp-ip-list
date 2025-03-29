resource "google_compute_network" "default" {
  name                    = "public-ip-list-network"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "proxy_subnet" {
  name          = "proxy-subnet"
  ip_cidr_range = "10.0.1.0/24"
  region        = var.region
  purpose       = "REGIONAL_MANAGED_PROXY"
  role          = "ACTIVE"
  network       = google_compute_network.default.id
}

resource "google_compute_subnetwork" "proxy_subnet_global" {
  name          = "proxy-subnet-global"
  ip_cidr_range = "10.0.4.0/24"
  purpose       = "GLOBAL_MANAGED_PROXY"
  role          = "ACTIVE"
  network       = google_compute_network.default.id
}

resource "google_compute_subnetwork" "lb_subnet" {
  name          = "lb-subnet"
  ip_cidr_range = "10.0.2.0/24"
  region        = var.region
  network       = google_compute_network.default.id
}

resource "google_compute_subnetwork" "backend_subnet" {
  name          = "backend-subnet"
  ip_cidr_range = "10.0.3.0/24"
  region        = var.region
  network       = google_compute_network.default.id
}