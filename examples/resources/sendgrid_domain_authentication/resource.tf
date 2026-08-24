resource "sendgrid_domain_authentication" "example" {
  domain             = "example.com"
  automatic_security = true
  default            = true
}

# After creating this resource, create the DNS records it reports with your
# DNS provider. SendGrid validates them asynchronously.
output "sendgrid_dns_records" {
  value = sendgrid_domain_authentication.example.dns_records
}
