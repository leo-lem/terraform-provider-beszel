package provider_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/leo-lem/terraform-provider-beszel/internal/provider"
)

// testAccProtoV6ProviderFactories is shared by every acceptance test in this package.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"beszel": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// testAccPreCheck verifies the environment variables needed to run acceptance
// tests against a real Beszel hub are set. Acceptance tests are otherwise
// skipped by terraform-plugin-testing unless TF_ACC=1 is set.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	for _, envVar := range []string{"BESZEL_ENDPOINT", "BESZEL_EMAIL", "BESZEL_PASSWORD", "BESZEL_TEST_USER_ID"} {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for acceptance tests", envVar)
		}
	}
}
