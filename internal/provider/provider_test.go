package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate the provider
// during acceptance testing.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"sendgrid": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates the required environment variables are set
// before running acceptance tests. Acceptance tests make real calls against
// the SendGrid API and are only run when TF_ACC=1 is set (enforced by
// resource.Test itself); this additionally ensures credentials are present.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("SENDGRID_API_KEY") == "" {
		t.Fatal("SENDGRID_API_KEY must be set for acceptance tests")
	}
}

var _ provider.Provider = &SendgridProvider{}
