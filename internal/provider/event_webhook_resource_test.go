package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEventWebhookResource(t *testing.T) {
	url := fmt.Sprintf("https://example.com/tf-acc-test-%d", testRandomSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sendgrid_event_webhook" "test" {
  url       = %[1]q
  enabled   = true
  delivered = true
  bounce    = true
}
`, url),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_event_webhook.test", "url", url),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.test", "delivered", "true"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.test", "bounce", "true"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.test", "open", "false"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.test", "signed", "false"),
					resource.TestCheckResourceAttrSet("sendgrid_event_webhook.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "sendgrid_event_webhook" "test" {
  url       = %[1]q
  enabled   = true
  delivered = true
  bounce    = true
  signed    = true
}
`, url),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_event_webhook.test", "signed", "true"),
					resource.TestCheckResourceAttrSet("sendgrid_event_webhook.test", "public_key"),
				),
			},
			{
				ResourceName:            "sendgrid_event_webhook.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oauth_client_secret"},
			},
		},
	})
}
