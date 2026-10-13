package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

// pocketBaseDateLayout is how PocketBase renders date fields.
const pocketBaseDateLayout = "2006-01-02 15:04:05.000Z"

// NewQuietHoursResource is the constructor registered with the provider.
func NewQuietHoursResource() resource.Resource {
	return &quietHoursResource{}
}

type quietHoursResource struct {
	client *client.Client
}

type quietHoursResourceModel struct {
	ID     types.String `tfsdk:"id"`
	User   types.String `tfsdk:"user"`
	System types.String `tfsdk:"system"`
	Type   types.String `tfsdk:"type"`
	Start  types.String `tfsdk:"start"`
	End    types.String `tfsdk:"end"`
}

func (r *quietHoursResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quiet_hours"
}

func (r *quietHoursResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a quiet hours window, during which alert notifications are muted for a user, " +
			"either for every system or for one system.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "PocketBase record ID of the quiet hours window.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user": schema.StringAttribute{
				Required:    true,
				Description: "ID of the user whose notifications are muted.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"system": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the system the window applies to. Omit to apply it to all systems.",
			},
			"type": schema.StringAttribute{
				Required: true,
				Description: "Kind of window. `one-time` mutes notifications from `start` (inclusive) until " +
					"`end` (exclusive), so `end` must be after `start`. `daily` repeats every day: only the " +
					"UTC time of day of `start` and `end` matters, the two must differ, and the window may " +
					"cross midnight (for example 22:00 to 06:00).",
				Validators: []validator.String{oneOf("one-time", "daily")},
			},
			"start": schema.StringAttribute{
				Required:    true,
				Description: "Start of the window, as an RFC 3339 timestamp.",
				Validators:  []validator.String{isRFC3339()},
			},
			"end": schema.StringAttribute{
				Required:    true,
				Description: "End of the window, as an RFC 3339 timestamp.",
				Validators:  []validator.String{isRFC3339()},
			},
		},
	}
}

// ValidateConfig rejects windows Beszel accepts but can never match, which
// would otherwise be silent no-ops. Beszel itself only validates the schema.
func (r *quietHoursResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg quietHoursResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if cfg.Type.IsNull() || cfg.Type.IsUnknown() || cfg.Start.IsUnknown() || cfg.End.IsUnknown() {
		return
	}
	start, errStart := time.Parse(time.RFC3339, cfg.Start.ValueString())
	end, errEnd := time.Parse(time.RFC3339, cfg.End.ValueString())
	if errStart != nil || errEnd != nil {
		return // reported by the attribute validators
	}

	switch cfg.Type.ValueString() {
	case "one-time":
		if !end.After(start) {
			resp.Diagnostics.AddAttributeError(path.Root("end"), "Invalid quiet hours window",
				"For a one-time window, end must be after start, otherwise the window never matches.")
		}
	case "daily":
		// Beszel compares only the UTC time of day, to the minute.
		if start.UTC().Format("15:04") == end.UTC().Format("15:04") {
			resp.Diagnostics.AddAttributeError(path.Root("end"), "Invalid quiet hours window",
				"For a daily window, start and end must differ in their UTC time of day (to the minute), "+
					"otherwise the window never matches.")
		}
	}
}

func (r *quietHoursResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *quietHoursResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan quietHoursResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	qh, err := r.client.CreateQuietHours(ctx, quietHoursWriteFromModel(plan, true))
	if err != nil {
		resp.Diagnostics.AddError("Error creating quiet hours", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, quietHoursModelFromAPI(qh, plan))...)
}

func (r *quietHoursResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state quietHoursResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	qh, err := r.client.GetQuietHours(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading quiet hours", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, quietHoursModelFromAPI(qh, state))...)
}

func (r *quietHoursResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state quietHoursResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	qh, err := r.client.UpdateQuietHours(ctx, state.ID.ValueString(), quietHoursWriteFromModel(plan, false))
	if err != nil {
		resp.Diagnostics.AddError("Error updating quiet hours", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, quietHoursModelFromAPI(qh, plan))...)
}

func (r *quietHoursResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state quietHoursResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteQuietHours(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting quiet hours", err.Error())
	}
}

func (r *quietHoursResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// quietHoursWriteFromModel builds the API payload. The user is only sent on
// create (it cannot change in place).
func quietHoursWriteFromModel(m quietHoursResourceModel, includeUser bool) client.QuietHoursWrite {
	w := client.QuietHoursWrite{
		System: m.System.ValueString(),
		Type:   m.Type.ValueString(),
		Start:  m.Start.ValueString(),
		End:    m.End.ValueString(),
	}
	if includeUser {
		w.User = m.User.ValueString()
	}
	return w
}

// quietHoursModelFromAPI maps an API record to the resource model. prior is
// the plan or state the record is being reconciled with: a timestamp keeps the
// user's own spelling when it denotes the same instant as the API value, so
// formatting and timezone offset differences never show up as drift.
func quietHoursModelFromAPI(qh *client.QuietHours, prior quietHoursResourceModel) quietHoursResourceModel {
	model := quietHoursResourceModel{
		ID:     types.StringValue(qh.ID),
		User:   types.StringValue(qh.User),
		System: types.StringNull(),
		Type:   types.StringValue(qh.Type),
		Start:  reconcileTimestamp(qh.Start, prior.Start),
		End:    reconcileTimestamp(qh.End, prior.End),
	}
	if qh.System != "" {
		model.System = types.StringValue(qh.System)
	}
	return model
}

func reconcileTimestamp(apiValue string, prior types.String) types.String {
	apiTime, err := time.Parse(pocketBaseDateLayout, apiValue)
	if err != nil {
		return types.StringValue(apiValue)
	}
	if !prior.IsNull() && !prior.IsUnknown() {
		if priorTime, err := time.Parse(time.RFC3339, prior.ValueString()); err == nil && priorTime.Equal(apiTime) {
			return prior
		}
	}
	return types.StringValue(apiTime.UTC().Format(time.RFC3339))
}
