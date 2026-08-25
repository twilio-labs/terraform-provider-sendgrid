resource "sendgrid_segment" "active_customers" {
  name = "active-customers"
  # query_dsl must be a full SELECT statement, not just a boolean condition —
  # confirmed against a live account. The SELECT clause must select exactly
  # contact_id and updated_at when querying the contact_data table.
  query_dsl = "SELECT contact_id, updated_at FROM contact_data WHERE email LIKE '%@example.com'"
}
