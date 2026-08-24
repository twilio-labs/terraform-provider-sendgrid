package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDomainAuthenticationResource(t *testing.T) {
	domain := fmt.Sprintf("tf-acc-test-%d.example.com", testRandomSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sendgrid_domain_authentication" "test" {
  domain             = %[1]q
  automatic_security = true
}
`, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_domain_authentication.test", "domain", domain),
					resource.TestCheckResourceAttrSet("sendgrid_domain_authentication.test", "id"),
					resource.TestCheckResourceAttrSet("sendgrid_domain_authentication.test", "dns_records.mail_cname_host"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "sendgrid_domain_authentication" "test" {
  domain             = %[1]q
  automatic_security = true
  default            = true
}
`, domain),
				Check: resource.TestCheckResourceAttr("sendgrid_domain_authentication.test", "default", "true"),
			},
			{
				ResourceName:      "sendgrid_domain_authentication.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
