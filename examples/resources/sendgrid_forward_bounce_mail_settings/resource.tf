resource "sendgrid_forward_bounce_mail_settings" "this" {
  enabled = true
  email   = "bounces@example.com"
}
