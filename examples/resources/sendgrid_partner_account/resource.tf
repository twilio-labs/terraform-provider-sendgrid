resource "sendgrid_partner_account" "customer" {
  profile = {
    first_name      = "Sender"
    last_name       = "Wiz"
    company_name    = "Example Co"
    company_website = "https://example.com"
    email           = "mail@example.com"
    timezone        = "Asia/Tokyo"
  }

  offerings = [
    {
      name     = "org.ei.free.v1"
      type     = "package"
      quantity = 1
    },
  ]

  state = "activated"
}
