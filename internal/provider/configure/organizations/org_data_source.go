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
	_ datasource.DataSource              = &organizationDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationDataSource{}
)

func NewOrganizationDataSource() datasource.DataSource {
	return &organizationDataSource{}
}

type organizationDataSource struct {
	client meraki.Client
}

type organizationDataSourceModel struct {
	ID                types.String   `tfsdk:"id"`
	Name              types.String   `tfsdk:"name"`
	APIEnabled        types.Bool     `tfsdk:"api_enabled"`
	LicensingModel    types.String   `tfsdk:"licensing_model"`
	CloudRegionName   types.String   `tfsdk:"cloud_region_name"`
	ManagementDetails []types.String `tfsdk:"management_details"`
}

func (o *organizationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org"
}

func (o *organizationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	o.client = client
	tflog.Info(ctx, "Configured the organizations data source")
}

func (o *organizationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier for the organization",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the organization",
			},
			"api_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Boolean indicating if the organization is API enabled",
			},
			"licensing_model": schema.StringAttribute{
				Computed:    true,
				Description: "The licensing model of the organization, can be 'co-term', 'per-device', or 'subscription'.",
			},
			"cloud_region_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the cloud region where the organization is located",
			},
			"management_details": schema.ListAttribute{
				Computed:    true,
				Description: "The management details of the organization, possibly empty. Details may be named 'MSP ID', 'IP restriction mode for API', or 'IP restriction mode for dashboard'.",
				ElementType: types.StringType,
			},
		},
	}
}

func (o *organizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan organizationDataSourceModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	organization, err := o.client.GetOrganization(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"failed to get organization",
			"failed to get organization: "+err.Error(),
		)
		return
	}

	var state organizationDataSourceModel
	state.ID = types.StringValue(organization.ID)
	state.Name = types.StringValue(organization.Name)
	state.APIEnabled = types.BoolValue(organization.API.Enabled)
	state.LicensingModel = types.StringValue(organization.Licensing.Model)
	state.CloudRegionName = types.StringValue(organization.Cloud.Region.Name)
	state.ManagementDetails = make([]types.String, 0, len(organization.Management.Details))
	for _, detail := range organization.Management.Details {
		state.ManagementDetails = append(state.ManagementDetails, types.StringValue(detail))
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
