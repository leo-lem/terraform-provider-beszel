package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

func TestStringPtrOrNil(t *testing.T) {
	if got := stringPtrOrNil(types.StringNull()); got != nil {
		t.Fatalf("expected nil for null, got %v", *got)
	}
	if got := stringPtrOrNil(types.StringUnknown()); got != nil {
		t.Fatalf("expected nil for unknown, got %v", *got)
	}
	got := stringPtrOrNil(types.StringValue("a"))
	if got == nil || *got != "a" {
		t.Fatalf("expected pointer to \"a\", got %v", got)
	}
}

func TestBoolPtrOrNil(t *testing.T) {
	if got := boolPtrOrNil(types.BoolNull()); got != nil {
		t.Fatalf("expected nil for null, got %v", *got)
	}
	if got := boolPtrOrNil(types.BoolUnknown()); got != nil {
		t.Fatalf("expected nil for unknown, got %v", *got)
	}
	got := boolPtrOrNil(types.BoolValue(true))
	if got == nil || *got != true {
		t.Fatalf("expected pointer to true, got %v", got)
	}
}

func TestFloat64PtrOrNil(t *testing.T) {
	if got := float64PtrOrNil(types.Float64Null()); got != nil {
		t.Fatalf("expected nil for null, got %v", *got)
	}
	got := float64PtrOrNil(types.Float64Value(90.5))
	if got == nil || *got != 90.5 {
		t.Fatalf("expected pointer to 90.5, got %v", got)
	}
}

func TestInt64PtrOrNil(t *testing.T) {
	if got := int64PtrOrNil(types.Int64Null()); got != nil {
		t.Fatalf("expected nil for null, got %v", *got)
	}
	got := int64PtrOrNil(types.Int64Value(10))
	if got == nil || *got != 10 {
		t.Fatalf("expected pointer to 10, got %v", got)
	}
}

func TestConfigureClient(t *testing.T) {
	var diags diag.Diagnostics
	if c := configureClient(nil, &diags); c != nil || diags.HasError() {
		t.Fatalf("expected nil client and no error for nil providerData, got client=%v diags=%v", c, diags)
	}

	diags = nil
	if c := configureClient("not a client", &diags); c != nil || !diags.HasError() {
		t.Fatalf("expected nil client and an error for wrong type, got client=%v diags=%v", c, diags)
	}

	diags = nil
	want := client.New("http://localhost:8090", "a@example.com", "password")
	got := configureClient(want, &diags)
	if got != want || diags.HasError() {
		t.Fatalf("expected the client to pass through unchanged, got %v diags=%v", got, diags)
	}
}

func TestStringOrEnv(t *testing.T) {
	const envVar = "BESZEL_TEST_STRING_OR_ENV"

	t.Run("config value wins", func(t *testing.T) {
		t.Setenv(envVar, "from-env")
		if got := stringOrEnv(types.StringValue("from-config"), envVar); got != "from-config" {
			t.Fatalf("expected config value, got %q", got)
		}
	})

	t.Run("falls back to env when null", func(t *testing.T) {
		t.Setenv(envVar, "from-env")
		if got := stringOrEnv(types.StringNull(), envVar); got != "from-env" {
			t.Fatalf("expected env value, got %q", got)
		}
	})

	t.Run("empty when neither set", func(t *testing.T) {
		t.Setenv(envVar, "")
		if got := stringOrEnv(types.StringNull(), envVar); got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})
}
