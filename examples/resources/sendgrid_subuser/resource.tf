resource "sendgrid_subuser" "app" {
  username = "my_app_subuser"
  email    = "subuser@example.com"
  password = var.subuser_password
  ips      = ["203.0.113.10"]
}

variable "subuser_password" {
  type      = string
  sensitive = true
}
