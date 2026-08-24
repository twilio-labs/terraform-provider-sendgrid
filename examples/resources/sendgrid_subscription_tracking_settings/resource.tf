resource "sendgrid_subscription_tracking_settings" "this" {
  enabled       = true
  plain_content = "If you would like to unsubscribe, click here: <% %>"
  html_content  = "<p>If you would like to unsubscribe, <% click here %></p>"
}
