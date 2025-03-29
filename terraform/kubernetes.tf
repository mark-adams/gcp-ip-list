resource "google_container_cluster" "default" {
  name     = "${local.prefix}-cluster"
  location = var.region

  remove_default_node_pool = true
  initial_node_count       = 1

  deletion_protection = false
}