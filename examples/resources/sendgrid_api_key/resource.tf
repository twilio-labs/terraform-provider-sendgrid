resource "sendgrid_api_key" "mail_send" {
  name   = "mail-send-key"
  scopes = ["mail.send"]
}
