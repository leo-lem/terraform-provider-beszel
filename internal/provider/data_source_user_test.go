package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "beszel" {}

resource "beszel_user" "test" {
  email    = "acc-test-ds@example.com"
  password = "acceptance-test-password-1"
  verified = true
}

data "beszel_user" "by_id" {
  id = beszel_user.test.id
}

data "beszel_user" "by_email" {
  email = beszel_user.test.email
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.beszel_user.by_id", "id", "beszel_user.test", "id"),
					resource.TestCheckResourceAttrPair("data.beszel_user.by_id", "email", "beszel_user.test", "email"),
					resource.TestCheckResourceAttrPair("data.beszel_user.by_email", "id", "beszel_user.test", "id"),
				),
			},
		},
	})
}
