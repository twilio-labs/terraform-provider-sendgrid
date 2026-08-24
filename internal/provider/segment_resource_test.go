package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSegmentResource(t *testing.T) {
	name := fmt.Sprintf("tf-acc-test-%d", testRandomSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sendgrid_segment" "test" {
  name      = %[1]q
  query_dsl = "email LIKE '%%@example.com'"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_segment.test", "name", name),
					resource.TestCheckResourceAttrSet("sendgrid_segment.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "sendgrid_segment" "test" {
  name      = %[1]q
  query_dsl = "email LIKE '%%@example.org'"
}
`, name+"-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_segment.test", "name", name+"-renamed"),
					resource.TestCheckResourceAttr("sendgrid_segment.test", "query_dsl", "email LIKE '%@example.org'"),
				),
			},
			{
				ResourceName:      "sendgrid_segment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
