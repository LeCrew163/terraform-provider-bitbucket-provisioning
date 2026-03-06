package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/LeCrew163/bitbucket-provisioning/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &projectsDataSource{}
	_ datasource.DataSourceWithConfigure = &projectsDataSource{}
)

func NewProjectsDataSource() datasource.DataSource {
	return &projectsDataSource{}
}

type projectsDataSource struct {
	client *client.Client
}

type projectsDataSourceModel struct {
	ID       types.String          `tfsdk:"id"`
	Filter   types.String          `tfsdk:"filter"`
	Projects []projectItemModel    `tfsdk:"projects"`
}

type projectItemModel struct {
	ID          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Public      types.Bool   `tfsdk:"public"`
}

func (d *projectsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

func (d *projectsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all Bitbucket Data Center projects accessible to the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Static identifier for this data source.",
				Computed:    true,
			},
			"filter": schema.StringAttribute{
				Description: "Optional substring to filter projects by name (case-insensitive).",
				Optional:    true,
			},
			"projects": schema.ListNestedAttribute{
				Description: "List of matching projects.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The project key (same as key).",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "The project key.",
							Computed:    true,
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
				},
			},
		},
	}
}

func (d *projectsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data projectsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filter := strings.ToLower(data.Filter.ValueString())
	tflog.Debug(ctx, "Reading projects data source", map[string]interface{}{"filter": filter})

	authCtx := d.client.NewAuthContext(ctx)
	const pageSize = float32(100)
	var start float32

	var projects []projectItemModel

	for {
		page, httpResp, err := d.client.GetAPIClient().ProjectAPI.
			GetProjects(authCtx).
			Start(start).
			Limit(pageSize).
			Execute()
		if err != nil {
			resp.Diagnostics.Append(client.HandleError("Failed to List Projects", client.ParseErrorResponse(httpResp))...)
			return
		}

		for _, p := range page.GetValues() {
			name := p.GetName()
			if filter != "" && !strings.Contains(strings.ToLower(name), filter) {
				continue
			}
			item := projectItemModel{
				ID:     types.StringValue(p.GetKey()),
				Key:    types.StringValue(p.GetKey()),
				Name:   types.StringValue(name),
				Public: types.BoolValue(p.GetPublic()),
			}
			if desc := p.GetDescription(); desc != "" {
				item.Description = types.StringValue(desc)
			} else {
				item.Description = types.StringNull()
			}
			projects = append(projects, item)
		}

		if page.GetIsLastPage() {
			break
		}
		start = float32(page.GetNextPageStart())
	}

	if projects == nil {
		projects = []projectItemModel{}
	}

	data.ID = types.StringValue("projects")
	data.Projects = projects
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
