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
	_ datasource.DataSource              = &repositoryDataSource{}
	_ datasource.DataSourceWithConfigure = &repositoryDataSource{}
)

// NewRepositoryDataSource is a helper function to simplify the provider implementation.
func NewRepositoryDataSource() datasource.DataSource {
	return &repositoryDataSource{}
}

type repositoryDataSource struct {
	client *client.Client
}

type repositoryDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectKey    types.String `tfsdk:"project_key"`
	Slug          types.String `tfsdk:"slug"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Public        types.Bool   `tfsdk:"public"`
	Forkable      types.Bool   `tfsdk:"forkable"`
	DefaultBranch types.String `tfsdk:"default_branch"`
	CloneURLHTTP  types.String `tfsdk:"clone_url_http"`
	CloneURLSSH   types.String `tfsdk:"clone_url_ssh"`
	State         types.String `tfsdk:"state"`
}

func (d *repositoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (d *repositoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Bitbucket Data Center repository by project key and slug.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier in the format {project_key}/{slug}.",
				Computed:    true,
			},
			"project_key": schema.StringAttribute{
				Description: "The key of the project that owns the repository.",
				Required:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The repository slug (URL-friendly name).",
				Required:    true,
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
			"state": schema.StringAttribute{
				Description: "The repository state (e.g. AVAILABLE).",
				Computed:    true,
			},
		},
	}
}

func (d *repositoryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *repositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data repositoryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.ProjectKey.ValueString()
	slug := data.Slug.ValueString()
	tflog.Debug(ctx, "Reading repository data source", map[string]interface{}{
		"project_key": projectKey,
		"slug":        slug,
	})

	authCtx := d.client.NewAuthContext(ctx)
	repo, httpResp, err := d.client.GetAPIClient().ProjectAPI.
		GetRepository(authCtx, projectKey, slug).Execute()
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddError(
				"Repository Not Found",
				fmt.Sprintf("No repository %q/%q was found.", projectKey, slug),
			)
			return
		}
		resp.Diagnostics.Append(client.HandleError("Failed to Read Repository", client.ParseErrorResponse(httpResp))...)
		return
	}

	repoSlug := repo.GetSlug()
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", projectKey, repoSlug))
	data.ProjectKey = types.StringValue(projectKey)
	data.Slug = types.StringValue(repoSlug)
	data.Name = types.StringValue(repo.GetName())
	data.Public = types.BoolValue(repo.GetPublic())
	data.Forkable = types.BoolValue(repo.GetForkable())
	data.State = types.StringValue(repo.GetState())
	data.DefaultBranch = types.StringValue(repo.GetDefaultBranch())

	if desc := repo.GetDescription(); desc != "" {
		data.Description = types.StringValue(desc)
	} else {
		data.Description = types.StringNull()
	}

	// Extract clone URLs from the links map.
	data.CloneURLHTTP = types.StringValue(extractCloneURL(repo.Links, "http"))
	data.CloneURLSSH = types.StringValue(extractCloneURL(repo.Links, "ssh"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
