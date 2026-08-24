resource "sendgrid_teammate" "analyst" {
  email    = "analyst@example.com"
  scopes   = ["user.profile.read", "stats.read"]
  is_admin = false
}
