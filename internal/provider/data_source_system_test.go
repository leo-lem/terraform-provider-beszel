package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSystemDataSource(t *testing.T) {
	userID := os.Getenv("BESZEL_TEST_USER_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "beszel" {}

resource "beszel_system" "test" {
  name  = "acc-test-system-ds"
  host  = "10.0.0.20"
  users = [%[1]q]
}

data "beszel_system" "by_id" {
  id = beszel_system.test.id
}

data "beszel_system" "by_name" {
  name = beszel_system.test.name
}
`, userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.beszel_system.by_id", "id", "beszel_system.test", "id"),
					resource.TestCheckResourceAttrPair("data.beszel_system.by_id", "host", "beszel_system.test", "host"),
					resource.TestCheckResourceAttrPair("data.beszel_system.by_name", "id", "beszel_system.test", "id"),
				),
			},
		},
	})
}

func TestAccSystemsDataSource(t *testing.T) {
	userID := os.Getenv("BESZEL_TEST_USER_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "beszel" {}

resource "beszel_system" "test" {
  name  = "acc-test-systems-ds"
  host  = "10.0.0.21"
  users = [%[1]q]
}

data "beszel_systems" "all" {
  filter = "id = \"${beszel_system.test.id}\""
}
`, userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.beszel_systems.all", "systems.#", "1"),
					resource.TestCheckResourceAttrPair("data.beszel_systems.all", "systems.0.id", "beszel_system.test", "id"),
				),
			},
		},
	})
}
