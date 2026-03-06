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
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

// NewUserDataSource is a helper function to simplify the provider implementation.
func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

type userDataSource struct {
	client *client.Client
}

type userDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Slug         types.String `tfsdk:"slug"`
	Name         types.String `tfsdk:"name"`
	DisplayName  types.String `tfsdk:"display_name"`
	EmailAddress types.String `tfsdk:"email_address"`
	Active       types.Bool   `tfsdk:"active"`
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Bitbucket Data Center user by slug (username).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The user slug (same as slug).",
				Computed:    true,
			},
			"slug": schema.StringAttribute{
				Description: "The username / slug of the user to look up.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The login name of the user.",
				Computed:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The full display name of the user.",
				Computed:    true,
			},
			"email_address": schema.StringAttribute{
				Description: "The email address of the user.",
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the user account is active.",
				Computed:    true,
			},
		},
	}
}

func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data userDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := data.Slug.ValueString()
	tflog.Debug(ctx, "Reading user data source", map[string]interface{}{"slug": slug})

	authCtx := d.client.NewAuthContext(ctx)
	user, httpResp, err := d.client.GetAPIClient().SystemMaintenanceAPI.
		GetUser(authCtx, slug).Execute()
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddError(
				"User Not Found",
				fmt.Sprintf("No user with slug %q was found.", slug),
			)
			return
		}
		resp.Diagnostics.Append(client.HandleError("Failed to Read User", client.ParseErrorResponse(httpResp))...)
		return
	}

	data.ID = types.StringValue(user.GetSlug())
	data.Slug = types.StringValue(user.GetSlug())
	data.Name = types.StringValue(user.GetName())
	data.DisplayName = types.StringValue(user.GetDisplayName())
	data.EmailAddress = types.StringValue(user.GetEmailAddress())
	data.Active = types.BoolValue(user.GetActive())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
