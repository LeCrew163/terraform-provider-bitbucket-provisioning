package provider

import (
	"context"
	"fmt"
	"net/http"

	"bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/internal/client"
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
	_ resource.Resource                = &projectPermissionsResource{}
	_ resource.ResourceWithConfigure   = &projectPermissionsResource{}
	_ resource.ResourceWithImportState = &projectPermissionsResource{}
)

// NewProjectPermissionsResource is a helper function to simplify the provider implementation.
func NewProjectPermissionsResource() resource.Resource {
	return &projectPermissionsResource{}
}

// projectPermissionsResource is the resource implementation.
type projectPermissionsResource struct {
	client *client.Client
}

// projectPermissionsResourceModel maps the resource schema data.
type projectPermissionsResourceModel struct {
	ID         types.String           `tfsdk:"id"`
	ProjectKey types.String           `tfsdk:"project_key"`
	Users      []userPermissionModel  `tfsdk:"user"`
	Groups     []groupPermissionModel `tfsdk:"group"`
}

type userPermissionModel struct {
	Name       types.String `tfsdk:"name"`
	Permission types.String `tfsdk:"permission"`
}

type groupPermissionModel struct {
	Name       types.String `tfsdk:"name"`
	Permission types.String `tfsdk:"permission"`
}

// Metadata returns the resource type name.
func (r *projectPermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_permissions"
}

// Schema defines the schema for the resource.
func (r *projectPermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	permValidator := &projectPermissionValidator{}

	resp.Schema = schema.Schema{
		Description: "Manages user and group permissions for a Bitbucket Data Center project. " +
			"This resource reconciles the desired set of permissions against the current state in Bitbucket, " +
			"adding, updating, and revoking permissions as needed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier (same as project_key).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_key": schema.StringAttribute{
				Description: "The key of the project to manage permissions for.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"user": schema.SetNestedBlock{
				Description: "A set of user permissions to grant on the project. " +
					"Users not listed here will have their project permissions revoked.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The Bitbucket username (slug).",
							Required:    true,
						},
						"permission": schema.StringAttribute{
							Description: "Permission level: PROJECT_READ, PROJECT_WRITE, or PROJECT_ADMIN.",
							Required:    true,
							Validators:  []validator.String{permValidator},
						},
					},
				},
			},
			"group": schema.SetNestedBlock{
				Description: "A set of group permissions to grant on the project. " +
					"Groups not listed here will have their project permissions revoked.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The Bitbucket group name.",
							Required:    true,
						},
						"permission": schema.StringAttribute{
							Description: "Permission level: PROJECT_READ, PROJECT_WRITE, or PROJECT_ADMIN.",
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
func (r *projectPermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *projectPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating project permissions", map[string]interface{}{
		"project_key": plan.ProjectKey.ValueString(),
	})

	resp.Diagnostics.Append(r.reconcilePermissions(ctx, plan.ProjectKey.ValueString(), plan.Users, plan.Groups)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = plan.ProjectKey
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the current permissions from Bitbucket.
func (r *projectPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading project permissions", map[string]interface{}{
		"project_key": state.ProjectKey.ValueString(),
	})

	projectKey := state.ProjectKey.ValueString()
	authCtx := r.client.NewAuthContext(ctx)

	users, err := r.listUserPermissions(authCtx, projectKey)
	if err != nil {
		if bbErr, ok := err.(*client.BitbucketError); ok && bbErr.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(client.HandleError("Failed to Read Project User Permissions", err)...)
		return
	}

	groups, err := r.listGroupPermissions(authCtx, projectKey)
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Read Project Group Permissions", err)...)
		return
	}

	state.ID = state.ProjectKey
	state.Users = users
	state.Groups = groups

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update reconciles the desired permissions against the current Bitbucket state.
func (r *projectPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating project permissions", map[string]interface{}{
		"project_key": plan.ProjectKey.ValueString(),
	})

	resp.Diagnostics.Append(r.reconcilePermissions(ctx, plan.ProjectKey.ValueString(), plan.Users, plan.Groups)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = plan.ProjectKey
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete revokes all permissions managed by this resource.
func (r *projectPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting project permissions (revoking all managed permissions)", map[string]interface{}{
		"project_key": state.ProjectKey.ValueString(),
	})

	// Reconcile to empty sets — this revokes all permissions that were managed by this resource.
	resp.Diagnostics.Append(r.reconcilePermissions(ctx, state.ProjectKey.ValueString(), nil, nil)...)
}

// ImportState imports the resource by project key.
func (r *projectPermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectKey := req.ID

	if !isValidProjectKey(projectKey) {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("The import ID %q is not a valid project key. "+
				"Project keys must be uppercase alphanumeric with underscores, 2-128 characters.", projectKey),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), projectKey)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_key"), projectKey)...)
}

// ── Reconciliation ───────────────────────────────────────────────────────────

// reconcilePermissions brings the project permissions in Bitbucket to match the desired state.
func (r *projectPermissionsResource) reconcilePermissions(
	ctx context.Context,
	projectKey string,
	desiredUsers []userPermissionModel,
	desiredGroups []groupPermissionModel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	authCtx := r.client.NewAuthContext(ctx)

	// ── Users ────────────────────────────────────────────────────────────────

	currentUsers, err := r.listUserPermissions(authCtx, projectKey)
	if err != nil {
		diags.Append(client.HandleError("Failed to List User Permissions", err)...)
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
			httpResp, setErr := r.client.GetAPIClient().ProjectAPI.
				SetPermissionForUsers1(authCtx, projectKey).
				Name(name).Permission(perm).Execute()
			if setErr != nil {
				diags.Append(client.HandleError(
					fmt.Sprintf("Failed to Grant User Permission (%s → %s)", name, perm),
					client.ParseErrorResponse(httpResp),
				)...)
				return diags
			}
		}
	}

	for name := range currentUserMap {
		if _, desired := desiredUserMap[name]; !desired {
			httpResp, revokeErr := r.client.GetAPIClient().ProjectAPI.
				RevokePermissionsForUser1(authCtx, projectKey).
				Name(name).Execute()
			if revokeErr != nil {
				diags.Append(client.HandleError(
					fmt.Sprintf("Failed to Revoke User Permission (%s)", name),
					client.ParseErrorResponse(httpResp),
				)...)
				return diags
			}
		}
	}

	// ── Groups ───────────────────────────────────────────────────────────────

	currentGroups, err := r.listGroupPermissions(authCtx, projectKey)
	if err != nil {
		diags.Append(client.HandleError("Failed to List Group Permissions", err)...)
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
			httpResp, setErr := r.client.GetAPIClient().ProjectAPI.
				SetPermissionForGroups1(authCtx, projectKey).
				Name(name).Permission(perm).Execute()
			if setErr != nil {
				diags.Append(client.HandleError(
					fmt.Sprintf("Failed to Grant Group Permission (%s → %s)", name, perm),
					client.ParseErrorResponse(httpResp),
				)...)
				return diags
			}
		}
	}

	for name := range currentGroupMap {
		if _, desired := desiredGroupMap[name]; !desired {
			httpResp, revokeErr := r.client.GetAPIClient().ProjectAPI.
				RevokePermissionsForGroup1(authCtx, projectKey).
				Name(name).Execute()
			if revokeErr != nil {
				diags.Append(client.HandleError(
					fmt.Sprintf("Failed to Revoke Group Permission (%s)", name),
					client.ParseErrorResponse(httpResp),
				)...)
				return diags
			}
		}
	}

	return diags
}

// listUserPermissions fetches all user permissions for a project, handling pagination.
func (r *projectPermissionsResource) listUserPermissions(
	authCtx context.Context,
	projectKey string,
) ([]userPermissionModel, error) {
	var result []userPermissionModel
	var start float32

	for {
		page, _, err := r.client.GetAPIClient().ProjectAPI.
			GetUsersWithAnyPermission1(authCtx, projectKey).
			Start(start).Limit(100).Execute()
		if err != nil {
			return nil, err
		}

		for _, pu := range page.Values {
			user := pu.GetUser()
			result = append(result, userPermissionModel{
				Name:       types.StringValue(user.Slug),
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

// listGroupPermissions fetches all group permissions for a project, handling pagination.
func (r *projectPermissionsResource) listGroupPermissions(
	authCtx context.Context,
	projectKey string,
) ([]groupPermissionModel, error) {
	var result []groupPermissionModel
	var start float32

	for {
		page, _, err := r.client.GetAPIClient().ProjectAPI.
			GetGroupsWithAnyPermission1(authCtx, projectKey).
			Start(start).Limit(100).Execute()
		if err != nil {
			return nil, err
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

// ── projectPermissionValidator ───────────────────────────────────────────────

type projectPermissionValidator struct{}

func (v *projectPermissionValidator) Description(_ context.Context) string {
	return "Permission must be one of: PROJECT_READ, PROJECT_WRITE, PROJECT_ADMIN"
}

func (v *projectPermissionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *projectPermissionValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	switch req.ConfigValue.ValueString() {
	case "PROJECT_READ", "PROJECT_WRITE", "PROJECT_ADMIN":
		// valid
	default:
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Permission Level",
			fmt.Sprintf("Permission %q is not valid. Must be one of: PROJECT_READ, PROJECT_WRITE, PROJECT_ADMIN",
				req.ConfigValue.ValueString()),
		)
	}
}
