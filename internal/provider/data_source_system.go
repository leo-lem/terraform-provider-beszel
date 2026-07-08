package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

// NewSystemDataSource is the constructor registered with the provider.
func NewSystemDataSource() datasource.DataSource {
	return &systemDataSource{}
}

type systemDataSource struct {
	client *client.Client
}

// systemDataSourceModel shares its shape with systemResourceModel (same
// tfsdk tags), so systemModelFromAPI in resource_system.go is reused here too.
type systemDataSourceModel = systemResourceModel

func (d *systemDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system"
}

func (d *systemDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a single existing Beszel system by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "PocketBase record ID of the system. Exactly one of id or name must be set.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Display name of the system. Exactly one of id or name must be set.",
			},
			"host": schema.StringAttribute{
				Computed:    true,
				Description: "Hostname or IP address the agent connects from.",
			},
			"port": schema.StringAttribute{
				Computed:    true,
				Description: "Agent port.",
			},
			"users": schema.ListAttribute{
				Computed:    true,
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

func (d *systemDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (d *systemDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config systemDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ID.IsUnknown() || config.Name.IsUnknown() {
		resp.Diagnostics.AddError(
			"Unknown value",
			"id and name must be known values; they cannot depend on an attribute that isn't known until apply.",
		)
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && config.Name.ValueString() != ""
	if hasID == hasName {
		resp.Diagnostics.AddError("Invalid configuration", "Exactly one of id or name must be set.")
		return
	}

	var system *client.System
	var err error
	if hasID {
		system, err = d.client.GetSystem(ctx, config.ID.ValueString())
	} else {
		var systems []client.System
		systems, err = d.client.ListSystems(ctx, fmt.Sprintf("name = %q", config.Name.ValueString()))
		if err == nil {
			if len(systems) == 0 {
				resp.Diagnostics.AddError("System not found", fmt.Sprintf("No system found with name %q.", config.Name.ValueString()))
				return
			}
			system = &systems[0]
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading system", err.Error())
		return
	}

	model := systemModelFromAPI(ctx, system, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
