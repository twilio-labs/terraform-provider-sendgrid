resource "sendgrid_template" "welcome_email" {
  name       = "welcome_email"
  generation = "dynamic"
}
