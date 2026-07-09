package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	ID     types.String `tfsdk:"id"`
	User   types.String `tfsdk:"user"`
	Emails types.List   `tfsdk:"emails"`
}

func (r *userSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_settings"
}

func (r *userSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a user's notification settings. Singleton per user.",
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

	var emails []string
	resp.Diagnostics.Append(plan.Emails.ElementsAs(ctx, &emails, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	existing, err := r.findByUser(ctx, userID)
	if err != nil {
		resp.Diagnostics.AddError("Error looking up existing user_settings", err.Error())
		return
	}

	write := client.UserSettingsWrite{
		User:     userID,
		Settings: client.UserSettingsData{Emails: emails},
	}

	settingsID := ""
	if existing != nil {
		settingsID = existing.ID
	} else {
		created, err := r.client.CreateUserSettings(ctx, write)
		if err != nil {
			resp.Diagnostics.AddError("Error creating user_settings", err.Error())
			return
		}
		settingsID = created.ID
	}

	// Beszel's OnRecordCreate("user_settings") hook always forces settings.emails
	// to the owning user's own address on create, discarding whatever was
	// posted for that field, so the desired emails always have to be applied
	// via a follow-up update, whether or not a record already existed.
	settings, err := r.client.UpdateUserSettings(ctx, settingsID, write)
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

	var emails []string
	resp.Diagnostics.Append(plan.Emails.ElementsAs(ctx, &emails, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.UpdateUserSettings(ctx, state.ID.ValueString(), client.UserSettingsWrite{
		Settings: client.UserSettingsData{Emails: emails},
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

func userSettingsModelFromAPI(ctx context.Context, settings *client.UserSettings) (userSettingsResourceModel, diag.Diagnostics) {
	emails, diags := types.ListValueFrom(ctx, types.StringType, settings.Settings.Emails)
	return userSettingsResourceModel{
		ID:     types.StringValue(settings.ID),
		User:   types.StringValue(settings.User),
		Emails: emails,
	}, diags
}
