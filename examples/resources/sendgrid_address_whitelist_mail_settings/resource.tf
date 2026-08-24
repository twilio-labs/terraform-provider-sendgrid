resource "sendgrid_address_whitelist_mail_settings" "this" {
  enabled = true
  list    = ["example.com", "trusted-sender@example.com"]
}
