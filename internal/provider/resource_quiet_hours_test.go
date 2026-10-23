package provider_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccQuietHoursResource(t *testing.T) {
	userID := os.Getenv("BESZEL_TEST_USER_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQuietHoursGlobalConfig(userID, "2026-01-01T22:00:00Z", "2026-01-02T06:00:00Z"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_quiet_hours.test", "type", "daily"),
					resource.TestCheckResourceAttr("beszel_quiet_hours.test", "start", "2026-01-01T22:00:00Z"),
					resource.TestCheckNoResourceAttr("beszel_quiet_hours.test", "system"),
					resource.TestCheckResourceAttrSet("beszel_quiet_hours.test", "id"),
				),
			},
			{
				ResourceName:      "beszel_quiet_hours.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQuietHoursGlobalConfig(userID, "2026-01-01T23:30:00Z", "2026-01-02T07:00:00Z"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_quiet_hours.test", "start", "2026-01-01T23:30:00Z"),
				),
			},
			{
				// Non-UTC offset on the same instants must not show as drift, and a
				// system scoped one-time window replaces the global daily one.
				Config: testAccQuietHoursSystemConfig(userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_quiet_hours.test", "type", "one-time"),
					resource.TestCheckResourceAttr("beszel_quiet_hours.test", "start", "2030-03-01T10:00:00+02:00"),
					resource.TestCheckResourceAttrPair("beszel_quiet_hours.test", "system", "beszel_system.test", "id"),
				),
			},
			{
				// Removing system goes back to applying to all systems. The system stays in
				// the config: Beszel cascade-deletes a system's quiet hours windows, so
				// deleting it in the same apply would race the update.
				Config: testAccQuietHoursGlobalConfig(userID, "2026-01-01T22:00:00Z", "2026-01-02T06:00:00Z") + testAccQuietHoursSystemOnly(userID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("beszel_quiet_hours.test", "system"),
				),
			},
		},
	})
}

func TestAccQuietHoursResourceInvalidWindow(t *testing.T) {
	userID := os.Getenv("BESZEL_TEST_USER_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccQuietHoursTypedConfig(userID, "one-time", "2030-03-01T12:00:00Z", "2030-03-01T10:00:00Z"),
				ExpectError: regexp.MustCompile(`end must be after start`),
			},
			{
				Config:      testAccQuietHoursTypedConfig(userID, "daily", "2030-03-01T10:00:00Z", "2030-03-02T10:00:30Z"),
				ExpectError: regexp.MustCompile(`must differ in their UTC time of day`),
			},
			{
				Config:      testAccQuietHoursTypedConfig(userID, "weekly", "2030-03-01T10:00:00Z", "2030-03-01T12:00:00Z"),
				ExpectError: regexp.MustCompile(`one-time, daily`),
			},
			{
				Config:      testAccQuietHoursTypedConfig(userID, "daily", "yesterday", "2030-03-01T12:00:00Z"),
				ExpectError: regexp.MustCompile(`RFC 3339`),
			},
		},
	})
}

func testAccQuietHoursTypedConfig(userID, typ, start, end string) string {
	return fmt.Sprintf(`
provider "beszel" {}

resource "beszel_quiet_hours" "test" {
  user  = %[1]q
  type  = %[2]q
  start = %[3]q
  end   = %[4]q
}
`, userID, typ, start, end)
}

func testAccQuietHoursGlobalConfig(userID, start, end string) string {
	return fmt.Sprintf(`
provider "beszel" {}

resource "beszel_quiet_hours" "test" {
  user  = %[1]q
  type  = "daily"
  start = %[2]q
  end   = %[3]q
}
`, userID, start, end)
}

func testAccQuietHoursSystemConfig(userID string) string {
	return fmt.Sprintf(`
provider "beszel" {}

resource "beszel_system" "test" {
  name  = "acc-test-quiet-hours-system"
  host  = "192.0.2.10"
  users = [%[1]q]
}

resource "beszel_quiet_hours" "test" {
  user   = %[1]q
  system = beszel_system.test.id
  type   = "one-time"
  start  = "2030-03-01T10:00:00+02:00"
  end    = "2030-03-01T12:00:00+02:00"
}
`, userID)
}

func testAccQuietHoursSystemOnly(userID string) string {
	return fmt.Sprintf(`
resource "beszel_system" "test" {
  name  = "acc-test-quiet-hours-system"
  host  = "192.0.2.10"
  users = [%[1]q]
}
`, userID)
}
