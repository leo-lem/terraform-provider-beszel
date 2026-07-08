package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserResourceConfig("acc-test@example.com", "user"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_user.test", "email", "acc-test@example.com"),
					resource.TestCheckResourceAttr("beszel_user.test", "role", "user"),
					resource.TestCheckResourceAttrSet("beszel_user.test", "id"),
					resource.TestCheckResourceAttrSet("beszel_user.test", "username"),
				),
			},
			{
				ResourceName:            "beszel_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: testAccUserResourceConfig("acc-test@example.com", "readonly"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_user.test", "role", "readonly"),
				),
			},
		},
	})
}

func testAccUserResourceConfig(email, role string) string {
	return fmt.Sprintf(`
provider "beszel" {}

resource "beszel_user" "test" {
  email    = %[1]q
  password = "acceptance-test-password-1"
  role     = %[2]q
  verified = true
}
`, email, role)
}
