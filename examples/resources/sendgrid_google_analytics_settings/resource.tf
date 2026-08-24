resource "sendgrid_google_analytics_settings" "this" {
  enabled      = true
  utm_source   = "sendgrid.com"
  utm_medium   = "email"
  utm_campaign = "spring-sale"
}
