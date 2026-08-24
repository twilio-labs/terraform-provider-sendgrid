resource "sendgrid_forward_spam_mail_settings" "this" {
  enabled = true
  email   = "spam-reports@example.com"
}
