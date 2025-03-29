resource "google_compute_address" "default" {
  name         = "${local.prefix}-public-not-used"
  address_type = "EXTERNAL"
  region       = google_compute_router.default.region
}
