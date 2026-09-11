terraform {
  required_providers {
    sendgrid = {
      source = "twilio-labs/sendgrid"
    }
  }
}

provider "sendgrid" {
  # api_key can also be set via the SENDGRID_API_KEY environment variable.
  api_key = var.sendgrid_api_key
}

variable "sendgrid_api_key" {
  type      = string
  sensitive = true
}
