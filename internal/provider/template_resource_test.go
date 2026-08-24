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

func TestParseTemplateVersionImportID(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		wantTemplateID string
		wantVersionID  string
		wantErr        bool
	}{
		{
			name:           "well-formed",
			id:             "tmpl-123/ver-456",
			wantTemplateID: "tmpl-123",
			wantVersionID:  "ver-456",
		},
		{
			name:    "missing separator",
			id:      "tmpl-123",
			wantErr: true,
		},
		{
			name:    "empty string",
			id:      "",
			wantErr: true,
		},
		{
			name:    "empty template_id segment",
			id:      "/ver-456",
			wantErr: true,
		},
		{
			name:    "empty version_id segment",
			id:      "tmpl-123/",
			wantErr: true,
		},
		{
			name:           "multiple slashes keep the rest in version_id",
			id:             "tmpl-123/ver-456/extra",
			wantTemplateID: "tmpl-123",
			wantVersionID:  "ver-456/extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templateID, versionID, err := parseTemplateVersionImportID(tt.id)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got templateID=%q versionID=%q", templateID, versionID)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if templateID != tt.wantTemplateID || versionID != tt.wantVersionID {
				t.Errorf("parseTemplateVersionImportID(%q) = (%q, %q), want (%q, %q)",
					tt.id, templateID, versionID, tt.wantTemplateID, tt.wantVersionID)
			}
		})
	}
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
