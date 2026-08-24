package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccClickTrackingSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "sendgrid_click_tracking_settings" "test" {
  enabled = true
}
`,
				Check: resource.TestCheckResourceAttr("sendgrid_click_tracking_settings.test", "enabled", "true"),
			},
			{
				Config: `
resource "sendgrid_click_tracking_settings" "test" {
  enabled = false
}
`,
				Check: resource.TestCheckResourceAttr("sendgrid_click_tracking_settings.test", "enabled", "false"),
			},
		},
	})
}

func TestAccOpenTrackingSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "sendgrid_open_tracking_settings" "test" {
  enabled = true
}
`,
				Check: resource.TestCheckResourceAttr("sendgrid_open_tracking_settings.test", "enabled", "true"),
			},
		},
	})
}

func TestAccSubscriptionTrackingSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "sendgrid_subscription_tracking_settings" "test" {
  enabled       = true
  plain_content = "Unsubscribe: <% %>"
  html_content  = "<p>Unsubscribe: <% %></p>"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_subscription_tracking_settings.test", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccGoogleAnalyticsSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "sendgrid_google_analytics_settings" "test" {
  enabled      = true
  utm_source   = "sendgrid.com"
  utm_medium   = "email"
  utm_campaign = "website"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_google_analytics_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("sendgrid_google_analytics_settings.test", "utm_source", "sendgrid.com"),
				),
			},
		},
	})
}
