package provider

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// stringOneOf validates that a string attribute is one of the allowed values.
type stringOneOf struct {
	allowed []string
}

func oneOf(allowed ...string) validator.String { return stringOneOf{allowed: allowed} }

func (v stringOneOf) Description(_ context.Context) string {
	return fmt.Sprintf("value must be one of: %s", strings.Join(v.allowed, ", "))
}

func (v stringOneOf) MarkdownDescription(ctx context.Context) string { return v.Description(ctx) }

func (v stringOneOf) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !slices.Contains(v.allowed, req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid value",
			fmt.Sprintf("%s, got: %q", v.Description(ctx), req.ConfigValue.ValueString()))
	}
}

// rfc3339 validates that a string attribute is an RFC 3339 timestamp.
type rfc3339 struct{}

func isRFC3339() validator.String { return rfc3339{} }

func (rfc3339) Description(_ context.Context) string {
	return "value must be an RFC 3339 timestamp, for example 2026-01-02T22:00:00Z"
}

func (v rfc3339) MarkdownDescription(ctx context.Context) string { return v.Description(ctx) }

func (v rfc3339) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if _, err := time.Parse(time.RFC3339, req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid timestamp",
			fmt.Sprintf("%s, got: %q", v.Description(ctx), req.ConfigValue.ValueString()))
	}
}

// shoutrrrURLs validates a list of Shoutrrr service URLs. It only catches the
// obvious mistakes (empty values, missing scheme://), not whether a given
// service scheme or its parameters are valid.
type shoutrrrURLs struct{}

func areShoutrrrURLs() validator.List { return shoutrrrURLs{} }

func (shoutrrrURLs) Description(_ context.Context) string {
	return "each value must be a Shoutrrr URL of the form scheme://..., for example generic://example.com/hook"
}

func (v shoutrrrURLs) MarkdownDescription(ctx context.Context) string { return v.Description(ctx) }

func (v shoutrrrURLs) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	for i, elem := range req.ConfigValue.Elements() {
		s, ok := elem.(types.String)
		if !ok || s.IsNull() || s.IsUnknown() {
			continue
		}
		if !validShoutrrrURL(s.ValueString()) {
			// The value itself is deliberately left out of the message: these
			// URLs usually embed credentials.
			resp.Diagnostics.AddAttributeError(req.Path.AtListIndex(i), "Invalid webhook URL", v.Description(ctx))
		}
	}
}

// shoutrrrURLPattern is deliberately loose: Shoutrrr URLs are not RFC 3986 URLs
// (for example slack://hook:T123-B456@webhook has a non-numeric "port"), so
// net/url cannot be used. Require a scheme, "://", and something after it.
var shoutrrrURLPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*://\S+$`)

func validShoutrrrURL(s string) bool { return shoutrrrURLPattern.MatchString(s) }
