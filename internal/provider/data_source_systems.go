package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

// NewSystemsDataSource is the constructor registered with the provider.
func NewSystemsDataSource() datasource.DataSource {
	return &systemsDataSource{}
}

type systemsDataSource struct {
	client *client.Client
}

type systemsDataSourceModel struct {
	Filter  types.String            `tfsdk:"filter"`
	Systems []systemDataSourceModel `tfsdk:"systems"`
}

func (d *systemsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_systems"
}

func (d *systemsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists Beszel systems, optionally filtered.",
		Attributes: map[string]schema.Attribute{
			"filter": schema.StringAttribute{
				Optional:    true,
				Description: "A PocketBase filter expression, e.g. status = \"up\". Leave unset to list all systems.",
			},
			"systems": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":     schema.StringAttribute{Computed: true},
						"name":   schema.StringAttribute{Computed: true},
						"host":   schema.StringAttribute{Computed: true},
						"port":   schema.StringAttribute{Computed: true},
						"users":  schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"status": schema.StringAttribute{Computed: true},
						"info":   schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *systemsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected data source configure type",
			fmt.Sprintf("Expected *client.Client, got %T. This is a provider bug.", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *systemsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config systemsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Filter.IsUnknown() {
		resp.Diagnostics.AddError(
			"Unknown value",
			"filter must be a known value; it cannot depend on an attribute that isn't known until apply.",
		)
		return
	}

	systems, err := d.client.ListSystems(ctx, config.Filter.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing systems", err.Error())
		return
	}

	models := make([]systemDataSourceModel, 0, len(systems))
	for i := range systems {
		model := systemDataSourceModelFromAPI(ctx, &systems[i], &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		models = append(models, model)
	}
	config.Systems = models

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
