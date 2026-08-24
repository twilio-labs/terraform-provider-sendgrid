package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccListResource(t *testing.T) {
	name := fmt.Sprintf("tf-acc-test-%d", testRandomSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sendgrid_list" "test" {
  name = %[1]q
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_list.test", "name", name),
					resource.TestCheckResourceAttr("sendgrid_list.test", "contact_count", "0"),
					resource.TestCheckResourceAttrSet("sendgrid_list.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "sendgrid_list" "test" {
  name = %[1]q
}
`, name+"-renamed"),
				Check: resource.TestCheckResourceAttr("sendgrid_list.test", "name", name+"-renamed"),
			},
			{
				ResourceName:      "sendgrid_list.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
