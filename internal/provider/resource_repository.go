package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/LeCrew163/bitbucket-provisioning/internal/client"
	bitbucket "github.com/LeCrew163/bitbucket-provisioning/internal/client/generated"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &repositoryResource{}
	_ resource.ResourceWithConfigure   = &repositoryResource{}
	_ resource.ResourceWithImportState = &repositoryResource{}
)

// NewRepositoryResource is a helper function to simplify the provider implementation.
func NewRepositoryResource() resource.Resource {
	return &repositoryResource{}
}

// repositoryResource is the resource implementation.
type repositoryResource struct {
	client *client.Client
}

// repositoryResourceModel maps the resource schema data.
type repositoryResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProjectKey     types.String `tfsdk:"project_key"`
	Slug           types.String `tfsdk:"slug"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Forkable       types.Bool   `tfsdk:"forkable"`
	Public         types.Bool   `tfsdk:"public"`
	DefaultBranch  types.String `tfsdk:"default_branch"`
	CloneURLHTTP   types.String `tfsdk:"clone_url_http"`
	CloneURLSSH    types.String `tfsdk:"clone_url_ssh"`
	State          types.String `tfsdk:"state"`
	PreventDestroy types.Bool   `tfsdk:"prevent_destroy"`
}

// Metadata returns the resource type name.
func (r *repositoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

// Schema defines the schema for the resource.
func (r *repositoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Bitbucket Data Center repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the repository, in the format {project_key}/{slug}.",
				Computed:    true,
			},
			"project_key": schema.StringAttribute{
				Description: "The key of the project containing this repository. Forces replacement if changed.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"slug": schema.StringAttribute{
				Description: "The repository slug (URL-friendly identifier). " +
					"Derived from the repository name by Bitbucket. Cannot be set directly.",
				Computed: true,
			},
			"name": schema.StringAttribute{
				Description: "The human-readable name of the repository.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the repository. " +
					"Computed because Bitbucket always returns a value for this field.",
				Optional: true,
				Computed: true,
			},
			"forkable": schema.BoolAttribute{
				Description: "Whether the repository can be forked. Defaults to true. " +
					"Computed because Bitbucket always returns a value for this field.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"public": schema.BoolAttribute{
				Description: "Whether the repository is publicly accessible. Defaults to false. " +
					"Computed because Bitbucket always returns a value for this field.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"default_branch": schema.StringAttribute{
				Description: "The default branch name. Populated once the repository has at least one commit.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"clone_url_http": schema.StringAttribute{
				Description: "The HTTP(S) clone URL.",
				Computed:    true,
			},
			"clone_url_ssh": schema.StringAttribute{
				Description: "The SSH clone URL.",
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "The repository state (e.g. AVAILABLE, INITIALISING).",
				Computed:    true,
			},
			"prevent_destroy": schema.BoolAttribute{
				Description: "When true (the default), Terraform will refuse to delete this repository. " +
					"Set to false explicitly to allow destruction.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *repositoryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (r *repositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan repositoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating repository", map[string]interface{}{
		"project_key": plan.ProjectKey.ValueString(),
		"name":        plan.Name.ValueString(),
	})

	repoReq := bitbucket.NewRestRepository()
	repoReq.SetName(plan.Name.ValueString())
	repoReq.SetScmId("git")

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		repoReq.SetDescription(plan.Description.ValueString())
	}
	if !plan.Forkable.IsNull() && !plan.Forkable.IsUnknown() {
		repoReq.SetForkable(plan.Forkable.ValueBool())
	}
	if !plan.Public.IsNull() && !plan.Public.IsUnknown() {
		repoReq.SetPublic(plan.Public.ValueBool())
	}

	authCtx := r.client.NewAuthContext(ctx)
	repo, httpResp, err := r.client.GetAPIClient().ProjectAPI.
		CreateRepository(authCtx, plan.ProjectKey.ValueString()).
		RestRepository(*repoReq).Execute()

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
			resp.Diagnostics.AddError(
				"Repository Already Exists",
				fmt.Sprintf("A repository derived from the name %q already exists in project %q. "+
					"Choose a different name or import the existing repository.",
					plan.Name.ValueString(), plan.ProjectKey.ValueString()),
			)
		} else {
			resp.Diagnostics.Append(client.HandleError("Failed to Create Repository", client.ParseErrorResponse(httpResp))...)
		}
		return
	}

	mapRepoToState(plan.ProjectKey.ValueString(), repo, &plan)

	// prevent_destroy is a local-only flag; preserve the planned value.
	if plan.PreventDestroy.IsNull() {
		plan.PreventDestroy = types.BoolValue(true)
	}

	tflog.Debug(ctx, "Repository created successfully", map[string]interface{}{
		"id": plan.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *repositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state repositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading repository", map[string]interface{}{
		"project_key": state.ProjectKey.ValueString(),
		"slug":        state.Slug.ValueString(),
	})

	authCtx := r.client.NewAuthContext(ctx)
	repo, httpResp, err := r.client.GetAPIClient().ProjectAPI.
		GetRepository(authCtx, state.ProjectKey.ValueString(), state.Slug.ValueString()).
		Execute()

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			tflog.Warn(ctx, "Repository not found, removing from state", map[string]interface{}{
				"project_key": state.ProjectKey.ValueString(),
				"slug":        state.Slug.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(client.HandleError("Failed to Read Repository", client.ParseErrorResponse(httpResp))...)
		return
	}

	mapRepoToState(state.ProjectKey.ValueString(), repo, &state)

	// prevent_destroy is not stored in Bitbucket; keep whatever is already in state.
	if state.PreventDestroy.IsNull() {
		state.PreventDestroy = types.BoolValue(true)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *repositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan repositoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read current state to get the existing slug for the API URL path.
	var state repositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	currentSlug := state.Slug.ValueString()

	tflog.Debug(ctx, "Updating repository", map[string]interface{}{
		"project_key":  plan.ProjectKey.ValueString(),
		"current_slug": currentSlug,
		"new_name":     plan.Name.ValueString(),
	})

	updateReq := bitbucket.NewRestRepository()
	updateReq.SetName(plan.Name.ValueString())
	// Always send description so users can clear it by setting it to "".
	updateReq.SetDescription(plan.Description.ValueString())
	if !plan.Forkable.IsNull() && !plan.Forkable.IsUnknown() {
		updateReq.SetForkable(plan.Forkable.ValueBool())
	}
	if !plan.Public.IsNull() && !plan.Public.IsUnknown() {
		updateReq.SetPublic(plan.Public.ValueBool())
	}

	authCtx := r.client.NewAuthContext(ctx)
	repo, httpResp, err := r.client.GetAPIClient().ProjectAPI.
		UpdateRepository(authCtx, plan.ProjectKey.ValueString(), currentSlug).
		RestRepository(*updateReq).Execute()

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddError(
				"Repository Not Found",
				fmt.Sprintf("Repository %q in project %q was not found.", currentSlug, plan.ProjectKey.ValueString()),
			)
		} else if httpResp != nil && httpResp.StatusCode == http.StatusForbidden {
			resp.Diagnostics.AddError(
				"Permission Denied",
				fmt.Sprintf("You don't have permission to update repository %q in project %q.", currentSlug, plan.ProjectKey.ValueString()),
			)
		} else {
			resp.Diagnostics.Append(client.HandleError("Failed to Update Repository", client.ParseErrorResponse(httpResp))...)
		}
		return
	}

	mapRepoToState(plan.ProjectKey.ValueString(), repo, &plan)

	// prevent_destroy is a local-only flag; preserve the planned value.
	if plan.PreventDestroy.IsNull() {
		plan.PreventDestroy = types.BoolValue(true)
	}

	tflog.Debug(ctx, "Repository updated successfully", map[string]interface{}{
		"id": plan.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *repositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state repositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.PreventDestroy.IsNull() || state.PreventDestroy.ValueBool() {
		resp.Diagnostics.AddError(
			"Repository Destruction Prevented",
			fmt.Sprintf(
				"Repository %q in project %q has prevent_destroy = true (the default). "+
					"Set prevent_destroy = false on this resource before running terraform destroy.",
				state.Slug.ValueString(), state.ProjectKey.ValueString(),
			),
		)
		return
	}

	tflog.Debug(ctx, "Deleting repository", map[string]interface{}{
		"project_key": state.ProjectKey.ValueString(),
		"slug":        state.Slug.ValueString(),
	})

	authCtx := r.client.NewAuthContext(ctx)
	httpResp, err := r.client.GetAPIClient().ProjectAPI.
		DeleteRepository(authCtx, state.ProjectKey.ValueString(), state.Slug.ValueString()).
		Execute()

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			tflog.Warn(ctx, "Repository already deleted", map[string]interface{}{
				"slug": state.Slug.ValueString(),
			})
			return
		}
		if httpResp != nil && httpResp.StatusCode == http.StatusForbidden {
			resp.Diagnostics.AddError(
				"Permission Denied",
				fmt.Sprintf("You don't have permission to delete repository %q.", state.Slug.ValueString()),
			)
		} else {
			resp.Diagnostics.Append(client.HandleError("Failed to Delete Repository", client.ParseErrorResponse(httpResp))...)
		}
		return
	}

	tflog.Debug(ctx, "Repository deleted successfully", map[string]interface{}{
		"slug": state.Slug.ValueString(),
	})
}

// ImportState imports the resource into Terraform state.
// Import ID format: {project_key}/{slug}
func (r *repositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in the format {project_key}/{slug}, got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_key"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("slug"), parts[1])...)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// mapRepoToState maps a Bitbucket RestRepository API response to the Terraform state model.
func mapRepoToState(projectKey string, repo *bitbucket.RestRepository, state *repositoryResourceModel) {
	state.ID = types.StringValue(projectKey + "/" + repo.GetSlug())
	state.ProjectKey = types.StringValue(projectKey)
	state.Slug = types.StringValue(repo.GetSlug())
	state.Name = types.StringValue(repo.GetName())

	// Normalize description: store as "" when absent so Optional+Computed works cleanly.
	if desc, ok := repo.GetDescriptionOk(); ok {
		state.Description = types.StringValue(*desc)
	} else {
		state.Description = types.StringValue("")
	}

	if forkable, ok := repo.GetForkableOk(); ok {
		state.Forkable = types.BoolValue(*forkable)
	} else {
		state.Forkable = types.BoolValue(true)
	}

	if public, ok := repo.GetPublicOk(); ok {
		state.Public = types.BoolValue(*public)
	} else {
		state.Public = types.BoolValue(false)
	}

	if branch, ok := repo.GetDefaultBranchOk(); ok {
		state.DefaultBranch = types.StringValue(*branch)
	} else {
		state.DefaultBranch = types.StringNull()
	}

	state.CloneURLHTTP = types.StringValue(extractCloneURL(repo.Links, "http"))
	state.CloneURLSSH = types.StringValue(extractCloneURL(repo.Links, "ssh"))

	if s, ok := repo.GetStateOk(); ok {
		state.State = types.StringValue(*s)
	} else {
		state.State = types.StringValue("")
	}
}

// extractCloneURL pulls the clone href for the given protocol name ("http" or "ssh")
// out of a Bitbucket repository links map.
func extractCloneURL(links map[string]interface{}, name string) string {
	if links == nil {
		return ""
	}
	cloneList, ok := links["clone"]
	if !ok {
		return ""
	}
	clones, ok := cloneList.([]interface{})
	if !ok {
		return ""
	}
	for _, c := range clones {
		entry, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if entry["name"] == name {
			if href, ok := entry["href"].(string); ok {
				return href
			}
		}
	}
	return ""
}

