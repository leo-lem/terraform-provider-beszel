package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

// NewUserResource is the constructor registered with the provider.
func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *client.Client
}

// userResourceModel maps the beszel_user schema to/from the API client's User type.
// Password is write-only: PocketBase never returns it, so it is not read back on refresh.
type userResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Email    types.String `tfsdk:"email"`
	Password types.String `tfsdk:"password"`
	Role     types.String `tfsdk:"role"`
	Verified types.Bool   `tfsdk:"verified"`
	Username types.String `tfsdk:"username"`
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Beszel user account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "PocketBase record ID of the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "Email address, used to sign in.",
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Password. PocketBase never returns this, so it cannot be read back or diffed against drift.",
			},
			"role": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("user"),
				Description: "One of user, admin, or readonly. Defaults to user.",
			},
			"verified": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the account is marked verified.",
			},
			"username": schema.StringAttribute{
				Computed:    true,
				Description: "Auto-generated username.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	password := plan.Password.ValueString()
	user, err := r.client.CreateUser(ctx, client.UserCreate{
		Email:           plan.Email.ValueString(),
		Password:        password,
		PasswordConfirm: password,
		Role:            plan.Role.ValueString(),
		Verified:        plan.Verified.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating user", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, userModelFromAPI(user, plan.Password))...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, userModelFromAPI(user, state.Password))...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	email := plan.Email.ValueString()
	role := plan.Role.ValueString()
	verified := plan.Verified.ValueBool()

	update := client.UserUpdate{
		Email:    &email,
		Role:     &role,
		Verified: &verified,
	}
	if !plan.Password.Equal(state.Password) {
		password := plan.Password.ValueString()
		update.Password = &password
		update.PasswordConfirm = &password
	}

	user, err := r.client.UpdateUser(ctx, state.ID.ValueString(), update)
	if err != nil {
		resp.Diagnostics.AddError("Error updating user", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, userModelFromAPI(user, plan.Password))...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteUser(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting user", err.Error())
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// userModelFromAPI converts an API User into the resource's Terraform model.
// password is passed through as-is since PocketBase never returns it.
func userModelFromAPI(user *client.User, password types.String) userResourceModel {
	return userResourceModel{
		ID:       types.StringValue(user.ID),
		Email:    types.StringValue(user.Email),
		Password: password,
		Role:     types.StringValue(user.Role),
		Verified: types.BoolValue(user.Verified),
		Username: types.StringValue(user.Username),
	}
}
