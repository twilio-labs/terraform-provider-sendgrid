resource "sendgrid_segment" "active_customers" {
  name      = "active-customers"
  query_dsl = "last_opened > '2024-01-01'"
}
