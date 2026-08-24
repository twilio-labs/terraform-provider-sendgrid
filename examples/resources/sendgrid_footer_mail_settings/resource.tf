resource "sendgrid_footer_mail_settings" "this" {
  enabled       = true
  html_content  = "<p>Sent with SendGrid</p>"
  plain_content = "Sent with SendGrid"
}
