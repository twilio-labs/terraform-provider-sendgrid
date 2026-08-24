resource "sendgrid_template_version" "welcome_email_v1" {
  template_id  = sendgrid_template.welcome_email.id
  name         = "v1"
  subject      = "Welcome, {{name}}!"
  html_content = "<p>Welcome, {{name}}!</p>"
  active       = true
}
