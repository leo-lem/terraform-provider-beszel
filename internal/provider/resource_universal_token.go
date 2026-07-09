package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

// NewUniversalTokenResource is the constructor registered with the provider.
func NewUniversalTokenResource() resource.Resource {
	return &universalTokenResource{}
}

type universalTokenResource struct {
	client *client.Client
}

// universalTokenResourceModel maps the beszel_universal_token schema to/from
// the API client's UniversalToken type. Singleton per user: Beszel scopes the
// underlying endpoint to one token per user, so user doubles as the resource's
// identity for Read/Import.
type universalTokenResourceModel struct {
	User      types.String `tfsdk:"user"`
	Token     types.String `tfsdk:"token"`
	Permanent types.Bool   `tfsdk:"permanent"`
	Active    types.Bool   `tfsdk:"active"`
}

func (r *universalTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_universal_token"
}

func (r *universalTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a user's universal agent registration token. Singleton per user: Beszel's " +
			"underlying endpoint is scoped to the caller's own record and forbids superuser callers, so the " +
			"provider impersonates the target user to manage it.",
		Attributes: map[string]schema.Attribute{
			"user": schema.StringAttribute{
				Required:    true,
				Description: "ID of the user this token belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Description: "Token value. Left unset, Beszel generates one.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"permanent": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the token persists across hub restarts. If false, it expires after an hour of hub uptime.",
			},
			"active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the token is currently active.",
			},
		},
	}
}

func (r *universalTokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *universalTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan universalTokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token, err := r.client.EnableUniversalToken(ctx, plan.User.ValueString(), plan.Token.ValueString(), plan.Permanent.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Error creating universal_token", err.Error())
		return
	}

	setState(ctx, &resp.State, &resp.Diagnostics, universalTokenModelFromAPI(token))
}

func (r *universalTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state universalTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token, err := r.client.GetUniversalToken(ctx, state.User.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading universal_token", err.Error())
		return
	}

	if !token.Active {
		resp.State.RemoveResource(ctx)
		return
	}

	setState(ctx, &resp.State, &resp.Diagnostics, universalTokenModelFromAPI(token))
}

func (r *universalTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan universalTokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token, err := r.client.EnableUniversalToken(ctx, plan.User.ValueString(), plan.Token.ValueString(), plan.Permanent.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Error updating universal_token", err.Error())
		return
	}

	setState(ctx, &resp.State, &resp.Diagnostics, universalTokenModelFromAPI(token))
}

func (r *universalTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state universalTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.DisableUniversalToken(ctx, state.User.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting universal_token", err.Error())
	}
}

func (r *universalTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("user"), req, resp)
}

func universalTokenModelFromAPI(token *client.UniversalToken) universalTokenResourceModel {
	return universalTokenResourceModel{
		User:      types.StringValue(token.User),
		Token:     types.StringValue(token.Token),
		Permanent: types.BoolValue(token.Permanent),
		Active:    types.BoolValue(token.Active),
	}
}
