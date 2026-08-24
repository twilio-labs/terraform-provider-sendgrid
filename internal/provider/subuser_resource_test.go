package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSubuserResource(t *testing.T) {
	username := fmt.Sprintf("tf_acc_test_%d", testRandomSuffix())
	email := username + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sendgrid_subuser" "test" {
  username = %[1]q
  email    = %[2]q
  password = "correct-horse-battery-staple-1!"
  ips      = ["127.0.0.1"]
}
`, username, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_subuser.test", "username", username),
					resource.TestCheckResourceAttr("sendgrid_subuser.test", "email", email),
					resource.TestCheckResourceAttr("sendgrid_subuser.test", "disabled", "false"),
					resource.TestCheckResourceAttrSet("sendgrid_subuser.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "sendgrid_subuser" "test" {
  username = %[1]q
  email    = %[2]q
  password = "correct-horse-battery-staple-1!"
  ips      = ["127.0.0.1"]
  disabled = true
}
`, username, email),
				Check: resource.TestCheckResourceAttr("sendgrid_subuser.test", "disabled", "true"),
			},
			{
				ResourceName:            "sendgrid_subuser.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}
