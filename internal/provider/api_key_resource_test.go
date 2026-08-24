package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAPIKeyResource(t *testing.T) {
	name := fmt.Sprintf("tf-acc-test-%d", testRandomSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sendgrid_api_key" "test" {
  name   = %[1]q
  scopes = ["mail.send"]
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_api_key.test", "name", name),
					resource.TestCheckResourceAttr("sendgrid_api_key.test", "scopes.#", "1"),
					resource.TestCheckTypeSetElemAttr("sendgrid_api_key.test", "scopes.*", "mail.send"),
					resource.TestCheckResourceAttrSet("sendgrid_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("sendgrid_api_key.test", "api_key"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "sendgrid_api_key" "test" {
  name   = %[1]q
  scopes = ["mail.send", "alerts.read"]
}
`, name+"-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_api_key.test", "name", name+"-renamed"),
					resource.TestCheckResourceAttr("sendgrid_api_key.test", "scopes.#", "2"),
				),
			},
			{
				ResourceName:      "sendgrid_api_key.test",
				ImportState:       true,
				ImportStateVerify: true,
				// api_key is never returned by the API after creation, so a
				// fresh import can't populate it.
				ImportStateVerifyIgnore: []string{"api_key"},
			},
		},
	})
}
