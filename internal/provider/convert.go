package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

// configureClient extracts the *client.Client passed down from the provider's
// Configure method, shared by every resource/data source Configure method.
// providerData is nil on the first call the framework makes before the
// provider itself has been configured, which is not an error.
func configureClient(providerData any, diags *diag.Diagnostics) *client.Client {
	if providerData == nil {
		return nil
	}

	c, ok := providerData.(*client.Client)
	if !ok {
		diags.AddError(
			"Unexpected configure type",
			fmt.Sprintf("Expected *client.Client, got %T. This is a provider bug.", providerData),
		)
		return nil
	}
	return c
}

// setState sets model into state unless diags already carries an error, e.g.
// from building model itself. Shared by every Create/Read/Update method to
// avoid repeating the "bail out before writing a possibly-incomplete model" check.
func setState[T any](ctx context.Context, state *tfsdk.State, diags *diag.Diagnostics, model T) {
	if diags.HasError() {
		return
	}
	diags.Append(state.Set(ctx, model)...)
}

// The following helpers convert Terraform Framework values to Go pointers,
// treating both null and unknown as "don't send this field" so partial
// updates never clobber a server-side value with an empty/zero one.

func stringPtrOrNil(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	value := v.ValueString()
	return &value
}

func boolPtrOrNil(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	value := v.ValueBool()
	return &value
}

func float64PtrOrNil(v types.Float64) *float64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	value := v.ValueFloat64()
	return &value
}

func int64PtrOrNil(v types.Int64) *int {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	value := int(v.ValueInt64())
	return &value
}
