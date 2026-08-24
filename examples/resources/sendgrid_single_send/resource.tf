resource "sendgrid_single_send" "spring_sale" {
  name = "spring-sale-2026"

  send_to = {
    list_ids = [sendgrid_list.newsletter.id]
  }

  email_config = {
    subject              = "Spring Sale is here!"
    html_content         = "<p>Save 20% this week only.</p>"
    suppression_group_id = 12345
  }

  # Omit send_at to leave this as an unscheduled draft. Setting it schedules
  # (or, with "now", immediately sends) the campaign.
  # send_at = "2026-03-01T09:00:00Z"
}
