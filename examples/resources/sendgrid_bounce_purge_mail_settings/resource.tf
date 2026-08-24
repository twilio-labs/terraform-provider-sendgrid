resource "sendgrid_bounce_purge_mail_settings" "this" {
  enabled      = true
  soft_bounces = 7
  hard_bounces = 30
}
