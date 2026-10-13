package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

// NewUserSettingsResource is the constructor registered with the provider.
func NewUserSettingsResource() resource.Resource {
	return &userSettingsResource{}
}

type userSettingsResource struct {
	client *client.Client
}

// userSettingsResourceModel maps the beszel_user_settings schema to/from the
// API client's UserSettings type. This is a singleton per user: Beszel enforces
// a unique index on the user field, so Create acts as an upsert against any
// record that may already exist (e.g. auto-created when the user first logged in).
type userSettingsResourceModel struct {
	ID       types.String `tfsdk:"id"`
	User     types.String `tfsdk:"user"`
	Emails   types.List   `tfsdk:"emails"`
	Webhooks types.List   `tfsdk:"webhooks"`
}

func (r *userSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_settings"
}

func (r *userSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a user's notification settings (email addresses and webhook / push notification " +
			"URLs). Singleton per user. Other keys Beszel stores in the same settings blob (units, " +
			"language, layout, ...) are left untouched.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "PocketBase record ID of the user_settings row.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user": schema.StringAttribute{
				Required:    true,
				Description: "ID of the user these settings belong to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"emails": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Email addresses that receive alert notifications for this user.",
			},
			"webhooks": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				Sensitive:   true,
				Validators:  []validator.List{areShoutrrrURLs()},
				Description: "Webhook / push notification URLs that receive alert notifications for this " +
					"user, in Shoutrrr URL format (for example telegram://token@telegram?chats=id or " +
					"generic://example.com/hook). See https://containrrr.dev/shoutrrr/. Treated as sensitive " +
					"because these URLs usually embed credentials.",
			},
		},
	}
}

func (r *userSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *userSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := plan.User.ValueString()

	emails, webhooks, diags := userSettingsListsFromPlan(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	existing, err := r.findByUser(ctx, userID)
	if err != nil {
		resp.Diagnostics.AddError("Error looking up existing user_settings", err.Error())
		return
	}

	current := existing
	if current == nil {
		created, err := r.client.CreateUserSettings(ctx, client.UserSettingsWrite{
			User:     userID,
			Settings: client.UserSettingsData{Emails: emails, Webhooks: webhooks},
		})
		if err != nil {
			resp.Diagnostics.AddError("Error creating user_settings", err.Error())
			return
		}
		current = created
	}

	// Beszel replaces the whole settings blob on write, so start from the
	// current one to keep keys this provider does not manage.
	write := client.UserSettingsWrite{
		User:     userID,
		Settings: current.Settings,
	}
	write.Settings.Emails = emails
	write.Settings.Webhooks = webhooks

	// Beszel's OnRecordCreate("user_settings") hook always forces settings.emails
	// to the owning user's own address on create, discarding whatever was
	// posted for that field, so the desired emails always have to be applied
	// via a follow-up update, whether or not a record already existed.
	settings, err := r.client.UpdateUserSettings(ctx, current.ID, write)
	if err != nil {
		resp.Diagnostics.AddError("Error setting user_settings", err.Error())
		return
	}

	model, diags := userSettingsModelFromAPI(ctx, settings)
	resp.Diagnostics.Append(diags...)
	setState(ctx, &resp.State, &resp.Diagnostics, model)
}

func (r *userSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.GetUserSettings(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading user_settings", err.Error())
		return
	}

	model, diags := userSettingsModelFromAPI(ctx, settings)
	resp.Diagnostics.Append(diags...)
	setState(ctx, &resp.State, &resp.Diagnostics, model)
}

func (r *userSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state userSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	emails, webhooks, diags := userSettingsListsFromPlan(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read-modify-write: Beszel replaces the whole settings blob on write.
	current, err := r.client.GetUserSettings(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading user_settings before update", err.Error())
		return
	}
	data := current.Settings
	data.Emails = emails
	data.Webhooks = webhooks

	settings, err := r.client.UpdateUserSettings(ctx, state.ID.ValueString(), client.UserSettingsWrite{
		Settings: data,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating user_settings", err.Error())
		return
	}

	model, diags := userSettingsModelFromAPI(ctx, settings)
	resp.Diagnostics.Append(diags...)
	setState(ctx, &resp.State, &resp.Diagnostics, model)
}

func (r *userSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteUserSettings(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting user_settings", err.Error())
	}
}

func (r *userSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// findByUser looks up an existing user_settings record for userID, returning
// nil (not an error) if none exists yet.
func (r *userSettingsResource) findByUser(ctx context.Context, userID string) (*client.UserSettings, error) {
	all, err := r.client.ListUserSettings(ctx, fmt.Sprintf("user = %q", userID))
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, nil
	}
	return &all[0], nil
}

func userSettingsListsFromPlan(ctx context.Context, plan userSettingsResourceModel) (emails, webhooks []string, diags diag.Diagnostics) {
	diags.Append(plan.Emails.ElementsAs(ctx, &emails, false)...)
	diags.Append(plan.Webhooks.ElementsAs(ctx, &webhooks, false)...)
	return emails, webhooks, diags
}

func userSettingsModelFromAPI(ctx context.Context, settings *client.UserSettings) (userSettingsResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	emails, d := types.ListValueFrom(ctx, types.StringType, nonNilStrings(settings.Settings.Emails))
	diags.Append(d...)
	webhooks, d := types.ListValueFrom(ctx, types.StringType, nonNilStrings(settings.Settings.Webhooks))
	diags.Append(d...)
	return userSettingsResourceModel{
		ID:       types.StringValue(settings.ID),
		User:     types.StringValue(settings.User),
		Emails:   emails,
		Webhooks: webhooks,
	}, diags
}

// nonNilStrings keeps absent API lists as empty lists rather than null.
func nonNilStrings(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}
