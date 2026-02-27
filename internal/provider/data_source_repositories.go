package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &repositoriesDataSource{}
	_ datasource.DataSourceWithConfigure = &repositoriesDataSource{}
)

func NewRepositoriesDataSource() datasource.DataSource {
	return &repositoriesDataSource{}
}

type repositoriesDataSource struct {
	client *client.Client
}

type repositoriesDataSourceModel struct {
	ID           types.String          `tfsdk:"id"`
	ProjectKey   types.String          `tfsdk:"project_key"`
	Filter       types.String          `tfsdk:"filter"`
	Repositories []repositoryItemModel `tfsdk:"repositories"`
}

type repositoryItemModel struct {
	ID            types.String `tfsdk:"id"`
	Slug          types.String `tfsdk:"slug"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Public        types.Bool   `tfsdk:"public"`
	State         types.String `tfsdk:"state"`
	Forkable      types.Bool   `tfsdk:"forkable"`
	DefaultBranch types.String `tfsdk:"default_branch"`
	CloneURLHTTP  types.String `tfsdk:"clone_url_http"`
	CloneURLSSH   types.String `tfsdk:"clone_url_ssh"`
}

func (d *repositoriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repositories"
}

func (d *repositoriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all repositories within a Bitbucket Data Center project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Static identifier for this data source in the format {project_key}/repositories.",
				Computed:    true,
			},
			"project_key": schema.StringAttribute{
				Description: "The key of the project to list repositories for.",
				Required:    true,
			},
			"filter": schema.StringAttribute{
				Description: "Optional substring to filter repositories by name (case-insensitive).",
				Optional:    true,
			},
			"repositories": schema.ListNestedAttribute{
				Description: "List of matching repositories.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Unique identifier in the format {project_key}/{slug}.",
							Computed:    true,
						},
						"slug": schema.StringAttribute{
							Description: "The repository slug.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The display name of the repository.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The repository description.",
							Computed:    true,
						},
						"public": schema.BoolAttribute{
							Description: "Whether the repository is publicly accessible.",
							Computed:    true,
						},
						"state": schema.StringAttribute{
							Description: "The repository state (e.g. AVAILABLE).",
							Computed:    true,
						},
						"forkable": schema.BoolAttribute{
							Description: "Whether the repository allows forking.",
							Computed:    true,
						},
						"default_branch": schema.StringAttribute{
							Description: "The default branch of the repository.",
							Computed:    true,
						},
						"clone_url_http": schema.StringAttribute{
							Description: "The HTTP clone URL.",
							Computed:    true,
						},
						"clone_url_ssh": schema.StringAttribute{
							Description: "The SSH clone URL.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *repositoriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *repositoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data repositoriesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.ProjectKey.ValueString()
	filter := strings.ToLower(data.Filter.ValueString())
	tflog.Debug(ctx, "Reading repositories data source", map[string]interface{}{
		"project_key": projectKey,
		"filter":      filter,
	})

	authCtx := d.client.NewAuthContext(ctx)
	const pageSize = float32(100)
	var start float32

	var repositories []repositoryItemModel

	for {
		page, httpResp, err := d.client.GetAPIClient().ProjectAPI.
			GetRepositories(authCtx, projectKey).
			Start(start).
			Limit(pageSize).
			Execute()
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddError(
					"Project Not Found",
					fmt.Sprintf("No project with key %q was found.", projectKey),
				)
				return
			}
			resp.Diagnostics.Append(client.HandleError("Failed to List Repositories", client.ParseErrorResponse(httpResp))...)
			return
		}

		for _, r := range page.GetValues() {
			name := r.GetName()
			if filter != "" && !strings.Contains(strings.ToLower(name), filter) {
				continue
			}
			item := repositoryItemModel{
				ID:            types.StringValue(fmt.Sprintf("%s/%s", projectKey, r.GetSlug())),
				Slug:          types.StringValue(r.GetSlug()),
				Name:          types.StringValue(name),
				Public:        types.BoolValue(r.GetPublic()),
				State:         types.StringValue(r.GetState()),
				Forkable:      types.BoolValue(r.GetForkable()),
				DefaultBranch: types.StringValue(r.GetDefaultBranch()),
				CloneURLHTTP:  types.StringValue(extractCloneURL(r.Links, "http")),
				CloneURLSSH:   types.StringValue(extractCloneURL(r.Links, "ssh")),
			}
			if desc := r.GetDescription(); desc != "" {
				item.Description = types.StringValue(desc)
			} else {
				item.Description = types.StringNull()
			}
			repositories = append(repositories, item)
		}

		if page.GetIsLastPage() {
			break
		}
		start = float32(page.GetNextPageStart())
	}

	if repositories == nil {
		repositories = []repositoryItemModel{}
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/repositories", projectKey))
	data.Repositories = repositories
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
