package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIPPoolResource(t *testing.T) {
	name := fmt.Sprintf("tf_acc_test_%d", testRandomSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sendgrid_ip_pool" "test" {
  name = %[1]q
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_ip_pool.test", "name", name),
					resource.TestCheckResourceAttrSet("sendgrid_ip_pool.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "sendgrid_ip_pool" "test" {
  name = %[1]q
}
`, name+"_renamed"),
				Check: resource.TestCheckResourceAttr("sendgrid_ip_pool.test", "name", name+"_renamed"),
			},
			{
				ResourceName:      "sendgrid_ip_pool.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
