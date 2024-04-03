package networks

import (
	"context"
	"fmt"
	"github.com/a60814billy/terraform-provider-cisco-meraki/meraki"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &networkResource{}
	_ resource.ResourceWithConfigure   = &networkResource{}
	_ resource.ResourceWithImportState = &networkResource{}
)

func NewNetworkResource() resource.Resource {
	return &networkResource{}
}

type networkResource struct {
	client meraki.Client
}

type NetworkResourceModel struct {
	ID           types.String `tfsdk:"id"`
	OrgID        types.String `tfsdk:"org_id"`
	Name         types.String `tfsdk:"name"`
	ProductTypes types.List   `tfsdk:"product_types"`
	TimeZone     types.String `tfsdk:"time_zone"`
	Tags         types.List   `tfsdk:"tags"`
	URL          types.String `tfsdk:"url"`
}

func (n *networkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (n *networkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the network",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL to the network Dashboard UI",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the organization to which the network belongs",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the network",
			},
			"product_types": schema.ListAttribute{
				Optional:    true,
				Description: "The product types of the network. Can be one or more of 'wireless', 'appliance', 'switch', 'systemsManager', 'camera', 'cellularGateway' or 'sensor'",
				ElementType: types.StringType,
			},
			"time_zone": schema.StringAttribute{
				Required:    true,
				Description: "The timezone of the network",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Description: "A list of tags to be applied to the network",
				ElementType: types.StringType,
			},
		},
	}
}

func (n *networkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Info(ctx, "Configuring the networks resource")
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

	n.client = client
	tflog.Info(ctx, "Configured the networks resourceresource")
}

func (n *networkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Creating the network resource")
	var plan NetworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "data", map[string]any{
		"org_id":   plan.OrgID.ValueString(),
		"name":     plan.Name.ValueString(),
		"timezone": plan.TimeZone.ValueString(),
	})

	//tfTags := make([]types.String, 0, len(plan.Tags.Elements()))
	//diags = plan.Tags.ElementsAs(ctx, &tfTags, false)
	//resp.Diagnostics.Append(diags...)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//tags := make([]string, 0, len(tfTags))
	//for _, tag := range tfTags {
	//	tags = append(tags, tag.ValueString())
	//}

	networkReqData := &meraki.NetworkCreateRequest{
		Name:         plan.Name.ValueString(),
		TimeZone:     plan.TimeZone.ValueString(),
		ProductTypes: []string{"systemsManager"},
	}

	network, err := n.client.CreateNetwork(plan.OrgID.ValueString(), networkReqData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create network",
			"Failed to create network: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "data", map[string]any{
		"network_id":     network.ID,
		"network_url":    network.URL,
		"network_name":   network.Name,
		"network_org_id": network.OrgID,
		"network_tz":     network.TimeZone,
	})

	var state NetworkResourceModel
	state.ID = types.StringValue(network.ID)
	state.OrgID = types.StringValue(network.OrgID)
	state.Name = types.StringValue(network.Name)
	state.TimeZone = types.StringValue(network.TimeZone)

	state.URL = types.StringValue(network.URL)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "Created the network resource")
}

func (n *networkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Reading the network resource")

	var state NetworkResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	network, err := n.client.GetNetwork(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get network",
			"Failed to get network: "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(network.ID)
	state.OrgID = types.StringValue(network.OrgID)
	state.Name = types.StringValue(network.Name)
	state.TimeZone = types.StringValue(network.TimeZone)

	var tfTags []attr.Value
	for _, tag := range network.Tags {
		tfTags = append(tfTags, types.StringValue(tag))
	}
	//state.Tags, diags = types.ListValue(types.StringType, tfTags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.URL = types.StringValue(network.URL)

	state.ProductTypes, diags = types.ListValueFrom(ctx, types.StringType, network.ProductTypes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Readed the network resource")
}

func (n *networkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Updating the network resource")
	var plan, state NetworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var networkUpdateReqData meraki.NetworkUpdateRequest

	// get changed fields
	if !plan.ProductTypes.Equal(state.ProductTypes) {
		resp.Diagnostics.AddError("Can't not change ProductTypes", "Can't not change ProductTypes")
		return
	}

	if !plan.Name.Equal(state.Name) {
		networkUpdateReqData.Name = plan.Name.ValueString()
	}
	if !plan.TimeZone.Equal(state.TimeZone) {
		networkUpdateReqData.TimeZone = plan.TimeZone.ValueString()
	}

	network, err := n.client.UpdateNetwork(state.ID.ValueString(), &networkUpdateReqData)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update network", "Failed to update network: "+err.Error())
		return
	}

	state.TimeZone = types.StringValue(network.TimeZone)
	state.Name = types.StringValue(network.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Upldated the network resource")
}

func (n *networkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Deleting the network resource")
	var state NetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := n.client.DeleteNetwork(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete network", "Failed to delete network: "+err.Error())
		return
	}
	tflog.Info(ctx, "Deleted the network resource")
}

func (n *networkResource) ImportState(ctx context.Context, request resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, resp)
}
