package provider

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestFilterAutoInjectedScopes(t *testing.T) {
	tests := []struct {
		name     string
		observed []string
		desired  []string
		want     []string
	}{
		{
			name:     "scope not in auto-injected map is always kept",
			observed: []string{"mail.send", "alerts.read"},
			desired:  []string{"mail.send", "alerts.read"},
			want:     []string{"mail.send", "alerts.read"},
		},
		{
			name:     "auto-injected scope absent from desired is dropped",
			observed: []string{"mail.send", "2fa_required"},
			desired:  []string{"mail.send"},
			want:     []string{"mail.send"},
		},
		{
			name:     "auto-injected scope explicitly requested is kept",
			observed: []string{"mail.send", "2fa_required"},
			desired:  []string{"mail.send", "2fa_required"},
			want:     []string{"mail.send", "2fa_required"},
		},
		{
			name:     "empty observed and desired",
			observed: nil,
			desired:  nil,
			want:     []string{},
		},
		{
			name: "all three known auto-injected scopes dropped together",
			observed: []string{
				"mail.send", "2fa_required", "sender_verification_eligible", "sender_verification_legacy",
			},
			desired: []string{"mail.send"},
			want:    []string{"mail.send"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterAutoInjectedScopes(tt.observed, tt.desired)
			sort.Strings(got)
			want := append([]string{}, tt.want...)
			sort.Strings(want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("filterAutoInjectedScopes(%v, %v) = %v, want %v", tt.observed, tt.desired, got, want)
			}
		})
	}
}

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
