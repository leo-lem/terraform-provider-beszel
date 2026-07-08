package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

// NewAlertResource is the constructor registered with the provider.
func NewAlertResource() resource.Resource {
	return &alertResource{}
}

type alertResource struct {
	client *client.Client
}

// alertResourceModel maps the beszel_alert schema to/from the API client's Alert type.
// The natural identity is (user, system, name), but the PocketBase id is what's stored
// and used for update/delete; user/system/name are ForceNew since changing any of them
// is really a different alert record.
type alertResourceModel struct {
	ID     types.String  `tfsdk:"id"`
	User   types.String  `tfsdk:"user"`
	System types.String  `tfsdk:"system"`
	Name   types.String  `tfsdk:"name"`
	Value  types.Float64 `tfsdk:"value"`
	Min    types.Int64   `tfsdk:"min"`

	Triggered types.Bool `tfsdk:"triggered"`
}

func (r *alertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

func (r *alertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an alert threshold for a system, on behalf of a user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "PocketBase record ID of the alert.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user": schema.StringAttribute{
				Required:    true,
				Description: "ID of the user the alert notifies.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"system": schema.StringAttribute{
				Required:    true,
				Description: "ID of the monitored system the alert applies to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				Description: "Kind of alert. One of Status, CPU, Memory, Disk, Temperature, " +
					"Bandwidth, GPU, LoadAvg1, LoadAvg5, LoadAvg15, or Battery.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.Float64Attribute{
				Optional:    true,
				Description: "Threshold value that triggers the alert.",
			},
			"min": schema.Int64Attribute{
				Optional:    true,
				Description: "Minutes the threshold must be exceeded before the alert triggers.",
			},
			"triggered": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the alert is currently in a triggered state.",
			},
		},
	}
}

func (r *alertResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource configure type",
			fmt.Sprintf("Expected *client.Client, got %T. This is a provider bug.", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *alertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alert, err := r.client.CreateAlert(ctx, client.AlertCreate{
		User:   plan.User.ValueString(),
		System: plan.System.ValueString(),
		Name:   plan.Name.ValueString(),
		Value:  float64PtrOrNil(plan.Value),
		Min:    int64PtrOrNil(plan.Min),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating alert", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, alertModelFromAPI(alert))...)
}

func (r *alertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alert, err := r.client.GetAlert(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading alert", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, alertModelFromAPI(alert))...)
}

func (r *alertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state alertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alert, err := r.client.UpdateAlert(ctx, state.ID.ValueString(), client.AlertUpdate{
		Value: float64PtrOrNil(plan.Value),
		Min:   int64PtrOrNil(plan.Min),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating alert", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, alertModelFromAPI(alert))...)
}

func (r *alertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAlert(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting alert", err.Error())
	}
}

func (r *alertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func alertModelFromAPI(alert *client.Alert) alertResourceModel {
	model := alertResourceModel{
		ID:        types.StringValue(alert.ID),
		User:      types.StringValue(alert.User),
		System:    types.StringValue(alert.System),
		Name:      types.StringValue(alert.Name),
		Triggered: types.BoolValue(alert.Triggered),
		Value:     types.Float64Null(),
		Min:       types.Int64Null(),
	}
	if alert.Value != nil {
		model.Value = types.Float64Value(*alert.Value)
	}
	if alert.Min != nil {
		model.Min = types.Int64Value(int64(*alert.Min))
	}
	return model
}
