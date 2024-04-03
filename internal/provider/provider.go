package provider

import (
	"context"
	"github.com/a60814billy/terraform-provider-cisco-meraki/internal/provider/configure/networks"
	"github.com/a60814billy/terraform-provider-cisco-meraki/internal/provider/configure/organizations"
	"github.com/a60814billy/terraform-provider-cisco-meraki/meraki"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ciscoMerakiProvider{
			version: version,
		}
	}
}

type ciscoMerakiProvider struct {
	version string
}

type ciscoMerakiProviderModel struct {
	APIKey string `tfsdk:"api_key"`
}

func (p *ciscoMerakiProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ciscomeraki"
	resp.Version = p.version
}

func (p *ciscoMerakiProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The API key used to authenticate with the Meraki API",
			},
		},
	}
}

func (p *ciscoMerakiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring the Cisco Meraki provider")
	var config ciscoMerakiProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := meraki.NewClient(config.APIKey)

	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured the Cisco Meraki provider")
}

func (p *ciscoMerakiProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		organizations.NewOrganizationsDataSource,
		organizations.NewOrganizationDataSource,
	}
}

func (p *ciscoMerakiProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		networks.NewNetworkResource,
	}
}
