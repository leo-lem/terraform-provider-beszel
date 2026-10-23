package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

// BeszelProvider implements provider.Provider.
type BeszelProvider struct {
	version string
}

// beszelProviderModel is the provider configuration block.
type beszelProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Email    types.String `tfsdk:"email"`
	Password types.String `tfsdk:"password"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BeszelProvider{version: version}
	}
}

func (p *BeszelProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "beszel"
	resp.Version = p.version
}

func (p *BeszelProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interacts with a Beszel hub. Authenticates as a Beszel superuser (PocketBase admin); " +
			"regular user auth cannot manage systems/users arbitrarily.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "Base URL of the Beszel hub, e.g. https://beszel.example.com. May also be set via the BESZEL_ENDPOINT environment variable. Required one way or the other.",
			},
			"email": schema.StringAttribute{
				Optional:    true,
				Description: "Email of a Beszel superuser account. May also be set via the BESZEL_EMAIL environment variable. Required one way or the other.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password of the Beszel superuser account. May also be set via the BESZEL_PASSWORD environment variable. Required one way or the other.",
			},
		},
	}
}

func (p *BeszelProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config beszelProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := stringOrEnv(config.Endpoint, "BESZEL_ENDPOINT")
	email := stringOrEnv(config.Email, "BESZEL_EMAIL")
	password := stringOrEnv(config.Password, "BESZEL_PASSWORD")

	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"), "Missing Beszel endpoint",
			"Set the endpoint attribute or the BESZEL_ENDPOINT environment variable.",
		)
	}
	if email == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("email"), "Missing Beszel email",
			"Set the email attribute or the BESZEL_EMAIL environment variable.",
		)
	}
	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"), "Missing Beszel password",
			"Set the password attribute or the BESZEL_PASSWORD environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	c := client.New(endpoint, email, password)
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *BeszelProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSystemResource,
		NewUserResource,
		NewAlertResource,
		NewUserSettingsResource,
		NewUniversalTokenResource,
		NewQuietHoursResource,
	}
}

func (p *BeszelProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSystemDataSource,
		NewSystemsDataSource,
		NewUserDataSource,
	}
}

func stringOrEnv(value types.String, envVar string) string {
	if !value.IsNull() && value.ValueString() != "" {
		return value.ValueString()
	}
	return os.Getenv(envVar)
}
