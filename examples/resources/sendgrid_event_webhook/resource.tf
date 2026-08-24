resource "sendgrid_event_webhook" "delivery_events" {
  url       = "https://example.com/webhooks/sendgrid"
  enabled   = true
  delivered = true
  bounce    = true
  dropped   = true
  signed    = true
}

# Use this to verify the signature on incoming requests.
output "sendgrid_webhook_public_key" {
  value = sendgrid_event_webhook.delivery_events.public_key
}
