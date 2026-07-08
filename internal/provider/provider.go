package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// BeszelProvider implements provider.Provider.
type BeszelProvider struct {
	version string
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
	// TODO: endpoint, email/password or token auth — see design notes in README.
}

func (p *BeszelProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
	// TODO: authenticate against the hub and inject a client into resources/data sources.
}

func (p *BeszelProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}

func (p *BeszelProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
