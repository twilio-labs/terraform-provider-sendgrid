resource "sendgrid_ip_pool" "marketing" {
  name = "marketing"
  ips  = ["203.0.113.10", "203.0.113.11"]
}
