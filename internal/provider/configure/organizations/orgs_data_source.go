package organizations

import (
	"context"
	"fmt"
	"github.com/a60814billy/terraform-provider-cisco-meraki/meraki"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &organizationsDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationsDataSource{}
)

func NewOrganizationsDataSource() datasource.DataSource {
	return &organizationsDataSource{}
}

type organizationsDataSource struct {
	client meraki.Client
}

type organizationsDataSourceModel struct {
	IDs []string `tfsdk:"ids"`
}

func (d *organizationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_orgs"
}

func (d *organizationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	tflog.Info(ctx, "Configuring the organizations data source")
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(meraki.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"invalid provider data",
			fmt.Sprintf("expected *meraki.Client, got %T. Please report this bug to the provider developer", req.ProviderData),
		)
		return
	}

	d.client = client
	tflog.Info(ctx, "Configured the organizations data source")
}

func (d *organizationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ids": schema.ListAttribute{
				Computed:    true,
				Description: "The unique identifiers for the organizations",
				ElementType: types.StringType,
			},
		},
	}
}

func (d *organizationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading the organizations data source")
	orgs, err := d.client.GetOrganizations()
	if err != nil {
		resp.Diagnostics.AddError("failed to get organizations", err.Error())
		return
	}

	var state organizationsDataSourceModel

	for _, org := range orgs {
		state.IDs = append(state.IDs, org.ID)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "Readed the organizations data source")
}
