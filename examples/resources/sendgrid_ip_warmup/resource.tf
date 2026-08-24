resource "sendgrid_ip_address" "dedicated" {
  is_auto_warmup     = false
  is_parent_assigned = true
}

resource "sendgrid_ip_warmup" "dedicated" {
  ip = sendgrid_ip_address.dedicated.ip
}
