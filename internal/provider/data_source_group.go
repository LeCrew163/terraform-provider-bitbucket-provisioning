package provider

import (
	"context"
	"fmt"

	"github.com/LeCrew163/bitbucket-provisioning/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &groupDataSource{}
	_ datasource.DataSourceWithConfigure = &groupDataSource{}
)

// NewGroupDataSource is a helper function to simplify the provider implementation.
func NewGroupDataSource() datasource.DataSource {
	return &groupDataSource{}
}

type groupDataSource struct {
	client *client.Client
}

type groupDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *groupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *groupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up an existing Bitbucket Data Center group by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The group name (same as name).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The exact name of the group to look up.",
				Required:    true,
			},
		},
	}
}

func (d *groupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	bitbucketClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = bitbucketClient
}

func (d *groupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data groupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	tflog.Debug(ctx, "Reading group data source", map[string]interface{}{"name": name})

	// Bitbucket does not provide a get-by-name endpoint; list with filter and
	// find the exact match.
	authCtx := d.client.NewAuthContext(ctx)
	page, httpResp, err := d.client.GetAPIClient().PermissionManagementAPI.
		GetGroups(authCtx).Filter(name).Limit(100).Execute()
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Read Group", client.ParseErrorResponse(httpResp))...)
		return
	}

	for _, g := range page.GetValues() {
		if g == name {
			data.ID = types.StringValue(name)
			data.Name = types.StringValue(name)
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"Group Not Found",
		fmt.Sprintf("No group with name %q was found. Check that the group exists in Bitbucket.", name),
	)
}
