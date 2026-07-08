package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

// NewSystemResource is the constructor registered with the provider.
func NewSystemResource() resource.Resource {
	return &systemResource{}
}

type systemResource struct {
	client *client.Client
}

// systemResourceModel maps the beszel_system schema to/from the API client's System type.
type systemResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Host  types.String `tfsdk:"host"`
	Port  types.String `tfsdk:"port"`
	Users types.List   `tfsdk:"users"`

	Status types.String `tfsdk:"status"`
	Info   types.String `tfsdk:"info"`
}

func (r *systemResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system"
}

func (r *systemResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a monitored system on a Beszel hub.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "PocketBase record ID of the system.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Display name of the system.",
			},
			"host": schema.StringAttribute{
				Required:    true,
				Description: "Hostname or IP address the agent connects from.",
			},
			"port": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Agent port. Left unset, Beszel assigns/reports one.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"users": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "IDs of Beszel users who own this system.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Agent-reported status: up, down, paused, or pending.",
			},
			"info": schema.StringAttribute{
				Computed:    true,
				Description: "Agent-reported system info, as a raw JSON string.",
			},
		},
	}
}

func (r *systemResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *systemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan systemResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var users []string
	resp.Diagnostics.Append(plan.Users.ElementsAs(ctx, &users, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	system, err := r.client.CreateSystem(ctx, client.SystemCreate{
		Name:  plan.Name.ValueString(),
		Host:  plan.Host.ValueString(),
		Port:  plan.Port.ValueString(),
		Users: users,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating system", err.Error())
		return
	}

	setState(ctx, &resp.State, &resp.Diagnostics, systemModelFromAPI(ctx, system, &resp.Diagnostics))
}

func (r *systemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state systemResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	system, err := r.client.GetSystem(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading system", err.Error())
		return
	}

	setState(ctx, &resp.State, &resp.Diagnostics, systemModelFromAPI(ctx, system, &resp.Diagnostics))
}

func (r *systemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state systemResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var users []string
	resp.Diagnostics.Append(plan.Users.ElementsAs(ctx, &users, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	host := plan.Host.ValueString()

	system, err := r.client.UpdateSystem(ctx, state.ID.ValueString(), client.SystemUpdate{
		Name:  &name,
		Host:  &host,
		Port:  stringPtrOrNil(plan.Port),
		Users: &users,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating system", err.Error())
		return
	}

	setState(ctx, &resp.State, &resp.Diagnostics, systemModelFromAPI(ctx, system, &resp.Diagnostics))
}

func (r *systemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state systemResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSystem(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting system", err.Error())
	}
}

func (r *systemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// systemModelFromAPI converts an API System into the resource's Terraform model.
// Diagnostics from converting the users list are appended to diags.
func systemModelFromAPI(ctx context.Context, system *client.System, diags *diag.Diagnostics) systemResourceModel {
	users, d := types.ListValueFrom(ctx, types.StringType, system.Users)
	diags.Append(d...)

	return systemResourceModel{
		ID:     types.StringValue(system.ID),
		Name:   types.StringValue(system.Name),
		Host:   types.StringValue(system.Host),
		Port:   types.StringValue(system.Port),
		Users:  users,
		Status: types.StringValue(system.Status),
		Info:   types.StringValue(string(system.Info)),
	}
}
