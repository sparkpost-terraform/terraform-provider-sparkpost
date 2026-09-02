package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccTrackingDomainResource exercises the full lifecycle of
// sparkpost_tracking_domain (create, update, read, import) against a real
// SparkPost account. It requires SPARKPOST_API_KEY (and TF_ACC=1) to run.
func TestAccTrackingDomainResource(t *testing.T) {
	domain := fmt.Sprintf("tf-acc-test-%d.example.com", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "sparkpost_tracking_domain" "test" {
  domain = %q
  https  = false
}
`, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sparkpost_tracking_domain.test", "domain", domain),
					resource.TestCheckResourceAttr("sparkpost_tracking_domain.test", "https", "false"),
				),
			},
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "sparkpost_tracking_domain" "test" {
  domain = %q
  https  = true
}
`, domain),
				Check: resource.TestCheckResourceAttr("sparkpost_tracking_domain.test", "https", "true"),
			},
			{
				ResourceName:      "sparkpost_tracking_domain.test",
				ImportState:       true,
				ImportStateId:     domain,
				ImportStateVerify: true,
			},
		},
	})
}
