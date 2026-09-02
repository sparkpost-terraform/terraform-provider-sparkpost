package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used to instantiate the provider during
// acceptance testing. The factory function is called for each Terraform CLI
// command executed by a test step.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"sparkpost": providerserver.NewProtocol6WithError(New()),
}

// testAccPreCheck validates the environment required for acceptance tests
// (SPARKPOST_API_KEY, and optionally SPARKPOST_API_URL) is configured. It is
// called before every acceptance test.
func testAccPreCheck(t *testing.T) {
	if os.Getenv("SPARKPOST_API_KEY") == "" {
		t.Fatal("SPARKPOST_API_KEY must be set for acceptance tests")
	}
}

// testAccProviderConfig returns a provider configuration block wired up from
// the SPARKPOST_API_KEY / SPARKPOST_API_URL environment variables.
func testAccProviderConfig() string {
	apiURL := os.Getenv("SPARKPOST_API_URL")
	if apiURL == "" {
		apiURL = "https://api.sparkpost.com/api/v1/"
	}

	return `
provider "sparkpost" {
  api_key = "` + os.Getenv("SPARKPOST_API_KEY") + `"
  api_url = "` + apiURL + `"
}
`
}
