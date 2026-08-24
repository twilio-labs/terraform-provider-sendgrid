resource "sendgrid_subuser" "marketing" {
  username = "marketing_subuser"
  email    = "marketing-subuser@example.com"
  password = var.subuser_password
  ips      = ["203.0.113.10"]
}

resource "sendgrid_ip_address" "dedicated" {
  is_auto_warmup     = true
  is_parent_assigned = false
  subusers           = [sendgrid_subuser.marketing.id]
}

variable "subuser_password" {
  type      = string
  sensitive = true
}
