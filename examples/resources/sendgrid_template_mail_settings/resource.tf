resource "sendgrid_template_mail_settings" "this" {
  enabled      = true
  html_content = "<% body %>"
}
