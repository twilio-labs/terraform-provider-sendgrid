package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTemplateResource(t *testing.T) {
	name := fmt.Sprintf("tf-acc-test-%d", testRandomSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sendgrid_template" "test" {
  name = %[1]q
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_template.test", "name", name),
					resource.TestCheckResourceAttr("sendgrid_template.test", "generation", "dynamic"),
					resource.TestCheckResourceAttrSet("sendgrid_template.test", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "sendgrid_template" "test" {
  name = %[1]q
}
`, name+"-renamed"),
				Check: resource.TestCheckResourceAttr("sendgrid_template.test", "name", name+"-renamed"),
			},
			{
				ResourceName:      "sendgrid_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// templateVersionImportStateIDFunc builds the composite "template_id/id"
// import identifier sendgrid_template_version expects.
func templateVersionImportStateIDFunc(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["sendgrid_template_version.test"]
	if !ok {
		return "", fmt.Errorf("resource not found in state: sendgrid_template_version.test")
	}
	return rs.Primary.Attributes["template_id"] + "/" + rs.Primary.Attributes["id"], nil
}

func TestAccTemplateVersionResource(t *testing.T) {
	templateName := fmt.Sprintf("tf-acc-test-%d", testRandomSuffix())
	versionName := "v1"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "sendgrid_template" "test" {
  name = %[1]q
}

resource "sendgrid_template_version" "test" {
  template_id  = sendgrid_template.test.id
  name         = %[2]q
  subject      = "Hello {{name}}"
  html_content = "<p>Hello {{name}}</p>"
  active       = true
}
`, templateName, versionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_template_version.test", "name", versionName),
					resource.TestCheckResourceAttr("sendgrid_template_version.test", "active", "true"),
					resource.TestCheckResourceAttrSet("sendgrid_template_version.test", "id"),
					resource.TestCheckResourceAttrSet("sendgrid_template_version.test", "plain_content"),
				),
			},
			{
				ResourceName:      "sendgrid_template_version.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: templateVersionImportStateIDFunc,
			},
		},
	})
}
