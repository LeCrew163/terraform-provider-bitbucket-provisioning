package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/LeCrew163/bitbucket-provisioning/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &repositoryPermissionsResource{}
	_ resource.ResourceWithConfigure   = &repositoryPermissionsResource{}
	_ resource.ResourceWithImportState = &repositoryPermissionsResource{}
)

// NewRepositoryPermissionsResource is a helper function to simplify the provider implementation.
func NewRepositoryPermissionsResource() resource.Resource {
	return &repositoryPermissionsResource{}
}

// repositoryPermissionsResource is the resource implementation.
type repositoryPermissionsResource struct {
	client *client.Client
}

// repositoryPermissionsResourceModel maps the resource schema data.
type repositoryPermissionsResourceModel struct {
	ID             types.String           `tfsdk:"id"`
	ProjectKey     types.String           `tfsdk:"project_key"`
	RepositorySlug types.String           `tfsdk:"repository_slug"`
	Users          []userPermissionModel  `tfsdk:"user"`
	Groups         []groupPermissionModel `tfsdk:"group"`
}

// Metadata returns the resource type name.
func (r *repositoryPermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_permissions"
}

// Schema defines the schema for the resource.
func (r *repositoryPermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	permValidator := &repoPermissionValidator{}

	resp.Schema = schema.Schema{
		Description: "Manages user and group permissions for a Bitbucket Data Center repository. " +
			"This resource reconciles the desired set of permissions against the current state in Bitbucket, " +
			"adding, updating, and revoking permissions as needed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier in the format {project_key}/{repository_slug}.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_key": schema.StringAttribute{
				Description: "The key of the project that owns the repository.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repository_slug": schema.StringAttribute{
				Description: "The slug of the repository to manage permissions for.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"user": schema.SetNestedBlock{
				Description: "A set of user permissions to grant on the repository. " +
					"Users not listed here will have their repository permissions revoked.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The Bitbucket username (slug).",
							Required:    true,
						},
						"permission": schema.StringAttribute{
							Description: "Permission level: REPO_READ, REPO_WRITE, or REPO_ADMIN.",
							Required:    true,
							Validators:  []validator.String{permValidator},
						},
					},
				},
			},
			"group": schema.SetNestedBlock{
				Description: "A set of group permissions to grant on the repository. " +
					"Groups not listed here will have their repository permissions revoked.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The Bitbucket group name.",
							Required:    true,
						},
						"permission": schema.StringAttribute{
							Description: "Permission level: REPO_READ, REPO_WRITE, or REPO_ADMIN.",
							Required:    true,
							Validators:  []validator.String{permValidator},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *repositoryPermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	bitbucketClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = bitbucketClient
}

// Create sets up permissions by reconciling the desired state against Bitbucket.
func (r *repositoryPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan repositoryPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating repository permissions", map[string]interface{}{
		"project_key":     plan.ProjectKey.ValueString(),
		"repository_slug": plan.RepositorySlug.ValueString(),
	})

	resp.Diagnostics.Append(r.reconcilePermissions(ctx, plan.ProjectKey.ValueString(), plan.RepositorySlug.ValueString(), plan.Users, plan.Groups)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.ProjectKey.ValueString(), plan.RepositorySlug.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the current permissions from Bitbucket.
func (r *repositoryPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state repositoryPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := state.ProjectKey.ValueString()
	repoSlug := state.RepositorySlug.ValueString()

	tflog.Debug(ctx, "Reading repository permissions", map[string]interface{}{
		"project_key":     projectKey,
		"repository_slug": repoSlug,
	})

	authCtx := r.client.NewAuthContext(ctx)

	users, err := r.listUserPermissions(authCtx, projectKey, repoSlug)
	if err != nil {
		if bbErr, ok := err.(*client.BitbucketError); ok && bbErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(client.HandleError("Failed to Read Repository User Permissions", err)...)
		return
	}

	groups, err := r.listGroupPermissions(authCtx, projectKey, repoSlug)
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Read Repository Group Permissions", err)...)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%s/%s", projectKey, repoSlug))
	state.Users = users
	state.Groups = groups

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update reconciles the desired permissions against the current Bitbucket state.
func (r *repositoryPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan repositoryPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating repository permissions", map[string]interface{}{
		"project_key":     plan.ProjectKey.ValueString(),
		"repository_slug": plan.RepositorySlug.ValueString(),
	})

	resp.Diagnostics.Append(r.reconcilePermissions(ctx, plan.ProjectKey.ValueString(), plan.RepositorySlug.ValueString(), plan.Users, plan.Groups)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.ProjectKey.ValueString(), plan.RepositorySlug.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete revokes all permissions managed by this resource.
func (r *repositoryPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state repositoryPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting repository permissions (revoking all managed permissions)", map[string]interface{}{
		"project_key":     state.ProjectKey.ValueString(),
		"repository_slug": state.RepositorySlug.ValueString(),
	})

	resp.Diagnostics.Append(r.reconcilePermissions(ctx, state.ProjectKey.ValueString(), state.RepositorySlug.ValueString(), nil, nil)...)
}

// ImportState imports the resource by project_key/repository_slug.
func (r *repositoryPermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in the format {project_key}/{repository_slug}, got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_key"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repository_slug"), parts[1])...)
}

// ── Reconciliation ───────────────────────────────────────────────────────────

func (r *repositoryPermissionsResource) reconcilePermissions(
	ctx context.Context,
	projectKey, repoSlug string,
	desiredUsers []userPermissionModel,
	desiredGroups []groupPermissionModel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	authCtx := r.client.NewAuthContext(ctx)

	// ── Users ────────────────────────────────────────────────────────────────

	currentUsers, err := r.listUserPermissions(authCtx, projectKey, repoSlug)
	if err != nil {
		diags.Append(client.HandleError("Failed to List Repository User Permissions", err)...)
		return diags
	}

	currentUserMap := make(map[string]string, len(currentUsers))
	for _, u := range currentUsers {
		currentUserMap[u.Name.ValueString()] = u.Permission.ValueString()
	}
	desiredUserMap := make(map[string]string, len(desiredUsers))
	for _, u := range desiredUsers {
		desiredUserMap[u.Name.ValueString()] = u.Permission.ValueString()
	}

	for name, perm := range desiredUserMap {
		if currentPerm, exists := currentUserMap[name]; !exists || currentPerm != perm {
			httpResp, setErr := r.client.GetAPIClient().PermissionManagementAPI.
				SetPermissionForUser(authCtx, projectKey, repoSlug).
				Name([]string{name}).Permission(perm).Execute()
			if setErr != nil {
				diags.Append(client.HandleError(
					fmt.Sprintf("Failed to Grant Repository User Permission (%s → %s)", name, perm),
					client.ParseErrorResponse(httpResp),
				)...)
				return diags
			}
		}
	}

	for name := range currentUserMap {
		if _, desired := desiredUserMap[name]; !desired {
			httpResp, revokeErr := r.client.GetAPIClient().PermissionManagementAPI.
				RevokePermissionsForUser2(authCtx, projectKey, repoSlug).
				Name(name).Execute()
			if revokeErr != nil {
				diags.Append(client.HandleError(
					fmt.Sprintf("Failed to Revoke Repository User Permission (%s)", name),
					client.ParseErrorResponse(httpResp),
				)...)
				return diags
			}
		}
	}

	// ── Groups ───────────────────────────────────────────────────────────────

	currentGroups, err := r.listGroupPermissions(authCtx, projectKey, repoSlug)
	if err != nil {
		diags.Append(client.HandleError("Failed to List Repository Group Permissions", err)...)
		return diags
	}

	currentGroupMap := make(map[string]string, len(currentGroups))
	for _, g := range currentGroups {
		currentGroupMap[g.Name.ValueString()] = g.Permission.ValueString()
	}
	desiredGroupMap := make(map[string]string, len(desiredGroups))
	for _, g := range desiredGroups {
		desiredGroupMap[g.Name.ValueString()] = g.Permission.ValueString()
	}

	for name, perm := range desiredGroupMap {
		if currentPerm, exists := currentGroupMap[name]; !exists || currentPerm != perm {
			httpResp, setErr := r.client.GetAPIClient().PermissionManagementAPI.
				SetPermissionForGroup(authCtx, projectKey, repoSlug).
				Name([]string{name}).Permission(perm).Execute()
			if setErr != nil {
				diags.Append(client.HandleError(
					fmt.Sprintf("Failed to Grant Repository Group Permission (%s → %s)", name, perm),
					client.ParseErrorResponse(httpResp),
				)...)
				return diags
			}
		}
	}

	for name := range currentGroupMap {
		if _, desired := desiredGroupMap[name]; !desired {
			httpResp, revokeErr := r.client.GetAPIClient().PermissionManagementAPI.
				RevokePermissionsForGroup2(authCtx, projectKey, repoSlug).
				Name(name).Execute()
			if revokeErr != nil {
				diags.Append(client.HandleError(
					fmt.Sprintf("Failed to Revoke Repository Group Permission (%s)", name),
					client.ParseErrorResponse(httpResp),
				)...)
				return diags
			}
		}
	}

	return diags
}

// listUserPermissions fetches all user permissions for a repository, handling pagination.
func (r *repositoryPermissionsResource) listUserPermissions(
	authCtx context.Context,
	projectKey, repoSlug string,
) ([]userPermissionModel, error) {
	var result []userPermissionModel
	var start float32

	for {
		page, httpResp, err := r.client.GetAPIClient().PermissionManagementAPI.
			GetUsersWithAnyPermission2(authCtx, projectKey, repoSlug).
			Start(start).Limit(100).Execute()
		if err != nil {
			return nil, client.ParseErrorResponse(httpResp)
		}

		for _, pu := range page.Values {
			user := pu.GetUser()
			result = append(result, userPermissionModel{
				Name:       types.StringValue(user.GetSlug()),
				Permission: types.StringValue(pu.GetPermission()),
			})
		}

		if page.GetIsLastPage() {
			break
		}
		start = float32(page.GetNextPageStart())
	}

	return result, nil
}

// listGroupPermissions fetches all group permissions for a repository, handling pagination.
func (r *repositoryPermissionsResource) listGroupPermissions(
	authCtx context.Context,
	projectKey, repoSlug string,
) ([]groupPermissionModel, error) {
	var result []groupPermissionModel
	var start float32

	for {
		page, httpResp, err := r.client.GetAPIClient().PermissionManagementAPI.
			GetGroupsWithAnyPermission2(authCtx, projectKey, repoSlug).
			Start(start).Limit(100).Execute()
		if err != nil {
			return nil, client.ParseErrorResponse(httpResp)
		}

		for _, pg := range page.Values {
			group := pg.GetGroup()
			result = append(result, groupPermissionModel{
				Name:       types.StringValue(group.GetName()),
				Permission: types.StringValue(pg.GetPermission()),
			})
		}

		if page.GetIsLastPage() {
			break
		}
		start = float32(page.GetNextPageStart())
	}

	return result, nil
}

// ── repoPermissionValidator ───────────────────────────────────────────────────

type repoPermissionValidator struct{}

func (v *repoPermissionValidator) Description(_ context.Context) string {
	return "Permission must be one of: REPO_READ, REPO_WRITE, REPO_ADMIN"
}

func (v *repoPermissionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *repoPermissionValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	switch req.ConfigValue.ValueString() {
	case "REPO_READ", "REPO_WRITE", "REPO_ADMIN":
		// valid
	default:
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Repository Permission Level",
			fmt.Sprintf("Permission %q is not valid. Must be one of: REPO_READ, REPO_WRITE, REPO_ADMIN",
				req.ConfigValue.ValueString()),
		)
	}
}
