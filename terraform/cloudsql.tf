resource "google_sql_database_instance" "default" {
  name             = "${local.prefix}-db"
  database_version = "POSTGRES_15"
  region           = var.region

  settings {
    tier = "db-f1-micro"

    ip_configuration {
      ipv4_enabled    = true
      private_network = google_compute_network.default.id
    }
  }

  depends_on = [google_service_networking_connection.cloudsql]

  deletion_protection = false
}

resource "google_compute_global_address" "cloudsql_private" {
  name          = "${local.prefix}-cloudsql-private"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.default.id
}

resource "google_service_networking_connection" "cloudsql" {
  network                 = google_compute_network.default.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.cloudsql_private.name]
}
