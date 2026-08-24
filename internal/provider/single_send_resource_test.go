package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSingleSendResource(t *testing.T) {
	name := fmt.Sprintf("tf-acc-test-%d", testRandomSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sendgrid_single_send" "test" {
  name = %[1]q

  send_to = {
    all = false
  }

  email_config = {
    subject              = "Hello"
    html_content         = "<p>Hello</p>"
    suppression_group_id = 1
  }
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_single_send.test", "name", name),
					resource.TestCheckResourceAttr("sendgrid_single_send.test", "status", "draft"),
					resource.TestCheckResourceAttrSet("sendgrid_single_send.test", "id"),
				),
			},
			{
				ResourceName:            "sendgrid_single_send.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"send_at"},
			},
		},
	})
}
