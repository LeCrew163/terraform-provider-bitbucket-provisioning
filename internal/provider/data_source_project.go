package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/LeCrew163/bitbucket-provisioning/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &projectDataSource{}
	_ datasource.DataSourceWithConfigure = &projectDataSource{}
)

// NewProjectDataSource is a helper function to simplify the provider implementation.
func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

type projectDataSource struct {
	client *client.Client
}

type projectDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Public      types.Bool   `tfsdk:"public"`
}

func (d *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *projectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Bitbucket Data Center project by key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The project key (same as key).",
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description: "The project key to look up (e.g. MYPROJ).",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The display name of the project.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The project description.",
				Computed:    true,
			},
			"public": schema.BoolAttribute{
				Description: "Whether the project is publicly accessible.",
				Computed:    true,
			},
		},
	}
}

func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data projectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.Key.ValueString()
	tflog.Debug(ctx, "Reading project data source", map[string]interface{}{"key": projectKey})

	authCtx := d.client.NewAuthContext(ctx)
	project, httpResp, err := d.client.GetAPIClient().ProjectAPI.
		GetProject(authCtx, projectKey).Execute()
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddError(
				"Project Not Found",
				fmt.Sprintf("No project with key %q was found.", projectKey),
			)
			return
		}
		resp.Diagnostics.Append(client.HandleError("Failed to Read Project", client.ParseErrorResponse(httpResp))...)
		return
	}

	data.ID = types.StringValue(project.GetKey())
	data.Key = types.StringValue(project.GetKey())
	data.Name = types.StringValue(project.GetName())

	if desc := project.GetDescription(); desc != "" {
		data.Description = types.StringValue(desc)
	} else {
		data.Description = types.StringNull()
	}

	data.Public = types.BoolValue(project.GetPublic())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
