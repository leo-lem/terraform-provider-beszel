package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAlertResource(t *testing.T) {
	userID := os.Getenv("BESZEL_TEST_USER_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertResourceConfig(userID, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_alert.test", "name", "CPU"),
					resource.TestCheckResourceAttr("beszel_alert.test", "min", "10"),
					resource.TestCheckResourceAttrSet("beszel_alert.test", "id"),
				),
			},
			{
				ResourceName:      "beszel_alert.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName: "beszel_alert.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["beszel_alert.test"]
					if !ok {
						return "", fmt.Errorf("beszel_alert.test not found in state")
					}
					attrs := rs.Primary.Attributes
					return fmt.Sprintf("%s/%s/%s", attrs["user"], attrs["system"], attrs["name"]), nil
				},
				ImportStateVerify: true,
			},
			{
				Config: testAccAlertResourceConfig(userID, 15),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_alert.test", "min", "15"),
				),
			},
		},
	})
}

func testAccAlertResourceConfig(userID string, min int) string {
	return fmt.Sprintf(`
provider "beszel" {}

resource "beszel_system" "test" {
  name  = "acc-test-alert-system"
  host  = "10.0.0.10"
  users = [%[1]q]
}

resource "beszel_alert" "test" {
  user   = %[1]q
  system = beszel_system.test.id
  name   = "CPU"
  value  = 90
  min    = %[2]d
}
`, userID, min)
}
