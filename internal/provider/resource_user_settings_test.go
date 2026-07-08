package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserSettingsResource(t *testing.T) {
	userID := os.Getenv("BESZEL_TEST_USER_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserSettingsResourceConfig(userID, "a@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_user_settings.test", "emails.#", "1"),
					resource.TestCheckResourceAttr("beszel_user_settings.test", "emails.0", "a@example.com"),
					resource.TestCheckResourceAttrSet("beszel_user_settings.test", "id"),
				),
			},
			{
				ResourceName:      "beszel_user_settings.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserSettingsResourceConfig(userID, "b@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_user_settings.test", "emails.0", "b@example.com"),
				),
			},
		},
	})
}

func testAccUserSettingsResourceConfig(userID, email string) string {
	return fmt.Sprintf(`
provider "beszel" {}

resource "beszel_user_settings" "test" {
  user   = %[1]q
  emails = [%[2]q]
}
`, userID, email)
}
