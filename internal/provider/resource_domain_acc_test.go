package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDomainResource exercises the full lifecycle of sparkpost_domain
// (create, read, import) against a real SparkPost account. It requires
// SPARKPOST_API_KEY (and TF_ACC=1) to run.
func TestAccDomainResource(t *testing.T) {
	domain := fmt.Sprintf("tf-acc-test-%d.example.com", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
resource "sparkpost_domain" "test" {
  domain = %q
}
`, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sparkpost_domain.test", "domain", domain),
					resource.TestCheckResourceAttr("sparkpost_domain.test", "id", domain),
				),
			},
			{
				ResourceName:      "sparkpost_domain.test",
				ImportState:       true,
				ImportStateId:     domain,
				ImportStateVerify: true,
			},
		},
	})
}
