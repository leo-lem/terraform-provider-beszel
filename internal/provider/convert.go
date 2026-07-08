package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

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
