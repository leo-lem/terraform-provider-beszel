package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSystemResource(t *testing.T) {
	userID := os.Getenv("BESZEL_TEST_USER_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSystemResourceConfig("acc-test-system", "10.0.0.1", userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_system.test", "name", "acc-test-system"),
					resource.TestCheckResourceAttr("beszel_system.test", "host", "10.0.0.1"),
					resource.TestCheckResourceAttrSet("beszel_system.test", "id"),
					resource.TestCheckResourceAttrSet("beszel_system.test", "status"),
				),
			},
			{
				ResourceName:      "beszel_system.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSystemResourceConfig("acc-test-system-renamed", "10.0.0.2", userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_system.test", "name", "acc-test-system-renamed"),
					resource.TestCheckResourceAttr("beszel_system.test", "host", "10.0.0.2"),
				),
			},
		},
	})
}

func testAccSystemResourceConfig(name, host, userID string) string {
	return fmt.Sprintf(`
provider "beszel" {}

resource "beszel_system" "test" {
  name  = %[1]q
  host  = %[2]q
  users = [%[3]q]
}
`, name, host, userID)
}
