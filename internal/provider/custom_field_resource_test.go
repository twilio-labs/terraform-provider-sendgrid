package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCustomFieldResource(t *testing.T) {
	name := fmt.Sprintf("tf_acc_test_%d", testRandomSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sendgrid_custom_field" "test" {
  name       = %[1]q
  field_type = "Text"
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_custom_field.test", "name", name),
					resource.TestCheckResourceAttr("sendgrid_custom_field.test", "field_type", "Text"),
					resource.TestCheckResourceAttrSet("sendgrid_custom_field.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "sendgrid_custom_field" "test" {
  name       = %[1]q
  field_type = "Text"
}
`, name+"_renamed"),
				Check: resource.TestCheckResourceAttr("sendgrid_custom_field.test", "name", name+"_renamed"),
			},
			{
				ResourceName:      "sendgrid_custom_field.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
