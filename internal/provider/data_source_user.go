package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

// NewUserDataSource is the constructor registered with the provider.
func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

type userDataSource struct {
	client *client.Client
}

type userDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Email types.String `tfsdk:"email"`

	Role     types.String `tfsdk:"role"`
	Verified types.Bool   `tfsdk:"verified"`
	Username types.String `tfsdk:"username"`
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a single existing Beszel user by ID or email.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "PocketBase record ID of the user. Exactly one of id or email must be set.",
			},
			"email": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Email address. Exactly one of id or email must be set.",
			},
			"role": schema.StringAttribute{
				Computed:    true,
				Description: "One of user, admin, or readonly.",
			},
			"verified": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the account is marked verified.",
			},
			"username": schema.StringAttribute{
				Computed:    true,
				Description: "Auto-generated username.",
			},
		},
	}
}

func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config userDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ID.IsUnknown() || config.Email.IsUnknown() {
		resp.Diagnostics.AddError(
			"Unknown value",
			"id and email must be known values; they cannot depend on an attribute that isn't known until apply.",
		)
		return
	}

	hasID := !config.ID.IsNull() && config.ID.ValueString() != ""
	hasEmail := !config.Email.IsNull() && config.Email.ValueString() != ""
	if hasID == hasEmail {
		resp.Diagnostics.AddError("Invalid configuration", "Exactly one of id or email must be set.")
		return
	}

	var user *client.User
	var err error
	if hasID {
		user, err = d.client.GetUser(ctx, config.ID.ValueString())
	} else {
		var users []client.User
		users, err = d.client.ListUsers(ctx, fmt.Sprintf("email = %q", config.Email.ValueString()))
		if err == nil {
			if len(users) == 0 {
				resp.Diagnostics.AddError("User not found", fmt.Sprintf("No user found with email %q.", config.Email.ValueString()))
				return
			}
			user = &users[0]
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, userDataSourceModel{
		ID:       types.StringValue(user.ID),
		Email:    types.StringValue(user.Email),
		Role:     types.StringValue(user.Role),
		Verified: types.BoolValue(user.Verified),
		Username: types.StringValue(user.Username),
	})...)
}
