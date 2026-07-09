package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUniversalTokenResource(t *testing.T) {
	userID := os.Getenv("BESZEL_TEST_USER_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUniversalTokenResourceConfig(userID, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_universal_token.test", "user", userID),
					resource.TestCheckResourceAttr("beszel_universal_token.test", "permanent", "false"),
					resource.TestCheckResourceAttr("beszel_universal_token.test", "active", "true"),
					resource.TestCheckResourceAttrSet("beszel_universal_token.test", "token"),
				),
			},
			{
				ResourceName: "beszel_universal_token.test",
				ImportState:  true,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return userID, nil
				},
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "user",
			},
			{
				Config: testAccUniversalTokenResourceConfig(userID, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_universal_token.test", "permanent", "true"),
					resource.TestCheckResourceAttr("beszel_universal_token.test", "active", "true"),
				),
			},
		},
	})
}

func testAccUniversalTokenResourceConfig(userID string, permanent bool) string {
	return fmt.Sprintf(`
provider "beszel" {}

resource "beszel_universal_token" "test" {
  user      = %[1]q
  permanent = %[2]t
}
`, userID, permanent)
}
