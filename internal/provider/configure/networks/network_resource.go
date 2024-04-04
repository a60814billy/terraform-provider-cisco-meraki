package networks

import (
	"context"
	"fmt"
	"github.com/a60814billy/terraform-provider-cisco-meraki/meraki"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
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
	ProductTypes types.Set    `tfsdk:"product_types"`
	TimeZone     types.String `tfsdk:"time_zone"`
	Tags         types.Set    `tfsdk:"tags"`
	URL          types.String `tfsdk:"url"`
	Notes        types.String `tfsdk:"notes"`
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
			"product_types": schema.SetAttribute{
				Optional:    true,
				Description: "The product types of the network. Can be one or more of 'wireless', 'appliance', 'switch', 'systemsManager', 'camera', 'cellularGateway' or 'sensor'",
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"time_zone": schema.StringAttribute{
				Required:    true,
				Description: "The timezone of the network",
			},
			"tags": schema.SetAttribute{
				Optional:    true,
				Description: "A list of tags to be applied to the network",
				ElementType: types.StringType,
			},
			"notes": schema.StringAttribute{
				Optional:    true,
				Description: "Add any notes or additional information about this network here",
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

	networkReqData := &meraki.NetworkCreateRequest{
		Name:     plan.Name.ValueString(),
		TimeZone: plan.TimeZone.ValueString(),
	}

	if !plan.Notes.IsNull() {
		networkReqData.Notes = plan.Notes.ValueString()
	}

	// set productTypes
	if plan.ProductTypes.IsNull() || len(plan.ProductTypes.Elements()) == 0 {
		// if productTypes is not set, set default values to all product types
		networkReqData.ProductTypes = []string{
			"wireless",
			"appliance",
			"switch",
			"camera",
			"cellularGateway",
			"sensor",
			"systemsManager",
		}
	} else if len(plan.ProductTypes.Elements()) > 0 {
		pts := make([]types.String, 0, len(plan.ProductTypes.Elements()))
		resp.Diagnostics.Append(plan.ProductTypes.ElementsAs(ctx, &pts, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, pt := range pts {
			networkReqData.ProductTypes = append(networkReqData.ProductTypes, pt.ValueString())
		}
	}

	if !plan.Tags.IsNull() && len(plan.Tags.Elements()) > 0 {
		tf_tags := make([]types.String, 0, len(plan.Tags.Elements()))
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tf_tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, tag := range tf_tags {
			networkReqData.Tags = append(networkReqData.Tags, tag.ValueString())
		}
	}

	tflog.Info(ctx, fmt.Sprintf("Data: %v", map[string]any{
		"productTypes": networkReqData.ProductTypes,
	}))

	network, err := n.client.CreateNetwork(plan.OrgID.ValueString(), networkReqData)
	tflog.Info(ctx, fmt.Sprintf("API called, Network: %v", network))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create network",
			fmt.Sprintf("Failed to create network: %s, %v", err.Error(), networkReqData),
		)
		return
	}

	plan.ID = types.StringValue(network.ID)
	plan.URL = types.StringValue(network.URL)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
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

	network, err := n.client.GetNetworkInOrg(state.OrgID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get network",
			"Failed to get network: "+err.Error(),
		)
		return
	}
	tflog.Info(ctx, fmt.Sprintf("Network: %v", network))

	state.ID = types.StringValue(network.ID)
	state.OrgID = types.StringValue(network.OrgID)

	if state.URL.IsNull() {
		// set URL only if it is not set
		// because it's changed frequently
		state.URL = types.StringValue(network.URL)
	}

	state.Name = types.StringValue(network.Name)
	state.TimeZone = types.StringValue(network.TimeZone)
	if len(network.Notes) > 0 {
		state.Notes = types.StringValue(network.Notes)
	}

	if len(network.ProductTypes) > 0 {
		tf_pts := make([]types.String, 0, len(network.ProductTypes))
		for _, pt := range network.ProductTypes {
			tf_pts = append(tf_pts, types.StringValue(pt))
		}
		pts, diags := types.SetValueFrom(ctx, types.StringType, tf_pts)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.ProductTypes = pts
	}

	if len(network.Tags) > 0 {
		tf_tags := make([]types.String, 0, len(network.Tags))
		for _, tag := range network.Tags {
			tf_tags = append(tf_tags, types.StringValue(tag))
		}
		tags, diags := types.SetValueFrom(ctx, types.StringType, tf_tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Tags = tags
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
	if !plan.Name.Equal(state.Name) {
		networkUpdateReqData.Name = plan.Name.ValueString()
	}
	if !plan.TimeZone.Equal(state.TimeZone) {
		networkUpdateReqData.TimeZone = plan.TimeZone.ValueString()
	}
	if !plan.Notes.Equal(state.Notes) {
		networkUpdateReqData.Notes = plan.Notes.ValueString()
	}
	if !plan.Tags.Equal(state.Tags) {
		tags := make([]types.String, 0, len(plan.Tags.Elements()))
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, tag := range tags {
			networkUpdateReqData.Tags = append(networkUpdateReqData.Tags, tag.ValueString())
		}
	}

	_, err := n.client.UpdateNetwork(state.ID.ValueString(), &networkUpdateReqData)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update network", "Failed to update network: "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
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
