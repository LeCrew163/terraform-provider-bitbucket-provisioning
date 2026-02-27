package provider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/internal/client"
	"bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/internal/client/generated"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

// NewProjectResource is a helper function to simplify the provider implementation.
func NewProjectResource() resource.Resource {
	return &projectResource{}
}

// projectResource is the resource implementation.
type projectResource struct {
	client *client.Client
}

// projectResourceModel maps the resource schema data.
type projectResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Key            types.String `tfsdk:"key"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Public         types.Bool   `tfsdk:"public"`
	PreventDestroy types.Bool   `tfsdk:"prevent_destroy"`
}

// Metadata returns the resource type name.
func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema defines the schema for the resource.
func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Bitbucket Data Center project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the project.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "The project key. Must be uppercase alphanumeric with underscores, 2-128 characters. " +
					"This is immutable and cannot be changed after creation.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					&projectKeyValidator{},
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the project.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the project.",
				Optional:    true,
			},
			"public": schema.BoolAttribute{
				Description: "Whether the project is public. Defaults to false (private). " +
					"Computed because Bitbucket always returns a value for this field.",
				Optional: true,
				Computed: true,
			},
			"prevent_destroy": schema.BoolAttribute{
				Description: "When true (the default), Terraform will refuse to delete this project. " +
					"Set to false explicitly to allow destruction.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating project", map[string]interface{}{
		"key":  plan.Key.ValueString(),
		"name": plan.Name.ValueString(),
	})

	// Build the project request
	projectReq := bitbucket.NewRestProject()
	projectReq.SetKey(plan.Key.ValueString())
	projectReq.SetName(plan.Name.ValueString())

	if !plan.Description.IsNull() {
		desc := plan.Description.ValueString()
		projectReq.SetDescription(desc)
	}

	if !plan.Public.IsNull() {
		projectReq.SetPublic(plan.Public.ValueBool())
	}

	// Create the project
	authCtx := r.client.NewAuthContext(ctx)
	project, httpResp, err := r.client.GetAPIClient().ProjectAPI.CreateProject(authCtx).RestProject(*projectReq).Execute()

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
			resp.Diagnostics.AddError(
				"Project Already Exists",
				fmt.Sprintf("A project with key '%s' already exists. Please choose a different key or import the existing project.", plan.Key.ValueString()),
			)
		} else {
			resp.Diagnostics.Append(client.HandleError("Failed to Create Project", client.ParseErrorResponse(httpResp))...)
		}
		return
	}

	// Map response to state
	plan.ID = types.StringValue(project.GetKey())
	plan.Key = types.StringValue(project.GetKey())
	plan.Name = types.StringValue(project.GetName())

	if desc, ok := project.GetDescriptionOk(); ok {
		plan.Description = types.StringValue(*desc)
	} else {
		plan.Description = types.StringNull()
	}

	if public, ok := project.GetPublicOk(); ok {
		plan.Public = types.BoolValue(*public)
	} else {
		plan.Public = types.BoolValue(false)
	}

	// prevent_destroy is a local-only flag; preserve the planned value.
	if plan.PreventDestroy.IsNull() {
		plan.PreventDestroy = types.BoolValue(true)
	}

	tflog.Debug(ctx, "Project created successfully", map[string]interface{}{
		"id":  plan.ID.ValueString(),
		"key": plan.Key.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading project", map[string]interface{}{
		"key": state.Key.ValueString(),
	})

	// Get the project
	authCtx := r.client.NewAuthContext(ctx)
	project, httpResp, err := r.client.GetAPIClient().ProjectAPI.GetProject(authCtx, state.Key.ValueString()).Execute()

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			// Project has been deleted outside Terraform
			tflog.Warn(ctx, "Project not found, removing from state", map[string]interface{}{
				"key": state.Key.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.Append(client.HandleError("Failed to Read Project", client.ParseErrorResponse(httpResp))...)
		return
	}

	// Update state
	state.ID = types.StringValue(project.GetKey())
	state.Key = types.StringValue(project.GetKey())
	state.Name = types.StringValue(project.GetName())

	if desc, ok := project.GetDescriptionOk(); ok {
		state.Description = types.StringValue(*desc)
	} else {
		state.Description = types.StringNull()
	}

	if public, ok := project.GetPublicOk(); ok {
		state.Public = types.BoolValue(*public)
	} else {
		state.Public = types.BoolValue(false)
	}

	// prevent_destroy is not stored in Bitbucket; keep whatever is already in state.
	if state.PreventDestroy.IsNull() {
		state.PreventDestroy = types.BoolValue(true)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating project", map[string]interface{}{
		"key":  plan.Key.ValueString(),
		"name": plan.Name.ValueString(),
	})

	// Build the update request
	updateReq := bitbucket.NewRestProject()
	updateReq.SetKey(plan.Key.ValueString())
	updateReq.SetName(plan.Name.ValueString())

	if !plan.Description.IsNull() {
		desc := plan.Description.ValueString()
		updateReq.SetDescription(desc)
	}

	if !plan.Public.IsNull() {
		updateReq.SetPublic(plan.Public.ValueBool())
	}

	// Update the project
	authCtx := r.client.NewAuthContext(ctx)
	project, httpResp, err := r.client.GetAPIClient().ProjectAPI.UpdateProject(authCtx, plan.Key.ValueString()).
		RestProject(*updateReq).Execute()

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddError(
				"Project Not Found",
				fmt.Sprintf("Project with key '%s' was not found. It may have been deleted outside of Terraform.", plan.Key.ValueString()),
			)
		} else if httpResp != nil && httpResp.StatusCode == http.StatusForbidden {
			resp.Diagnostics.AddError(
				"Permission Denied",
				fmt.Sprintf("You don't have permission to update project '%s'.", plan.Key.ValueString()),
			)
		} else {
			resp.Diagnostics.Append(client.HandleError("Failed to Update Project", client.ParseErrorResponse(httpResp))...)
		}
		return
	}

	// Update state
	plan.ID = types.StringValue(project.GetKey())
	plan.Key = types.StringValue(project.GetKey())
	plan.Name = types.StringValue(project.GetName())

	if desc, ok := project.GetDescriptionOk(); ok {
		plan.Description = types.StringValue(*desc)
	} else {
		plan.Description = types.StringNull()
	}

	if public, ok := project.GetPublicOk(); ok {
		plan.Public = types.BoolValue(*public)
	} else {
		plan.Public = types.BoolValue(false)
	}

	// prevent_destroy is a local-only flag; preserve the planned value.
	if plan.PreventDestroy.IsNull() {
		plan.PreventDestroy = types.BoolValue(true)
	}

	tflog.Debug(ctx, "Project updated successfully", map[string]interface{}{
		"key": plan.Key.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.PreventDestroy.IsNull() || state.PreventDestroy.ValueBool() {
		resp.Diagnostics.AddError(
			"Project Destruction Prevented",
			fmt.Sprintf(
				"Project %q has prevent_destroy = true (the default). "+
					"Set prevent_destroy = false on this resource before running terraform destroy.",
				state.Key.ValueString(),
			),
		)
		return
	}

	tflog.Debug(ctx, "Deleting project", map[string]interface{}{
		"key": state.Key.ValueString(),
	})

	// Delete the project
	authCtx := r.client.NewAuthContext(ctx)
	httpResp, err := r.client.GetAPIClient().ProjectAPI.DeleteProject(authCtx, state.Key.ValueString()).Execute()

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			// Project already deleted, consider it successful
			tflog.Warn(ctx, "Project already deleted", map[string]interface{}{
				"key": state.Key.ValueString(),
			})
			return
		}

		if httpResp != nil && httpResp.StatusCode == http.StatusForbidden {
			resp.Diagnostics.AddError(
				"Permission Denied",
				fmt.Sprintf("You don't have permission to delete project '%s'.", state.Key.ValueString()),
			)
		} else {
			resp.Diagnostics.Append(client.HandleError("Failed to Delete Project", client.ParseErrorResponse(httpResp))...)
		}
		return
	}

	tflog.Debug(ctx, "Project deleted successfully", map[string]interface{}{
		"key": state.Key.ValueString(),
	})
}

// ImportState imports the resource into Terraform state.
func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by project key
	projectKey := req.ID

	tflog.Debug(ctx, "Importing project", map[string]interface{}{
		"key": projectKey,
	})

	// Validate project key format
	if !isValidProjectKey(projectKey) {
		resp.Diagnostics.AddError(
			"Invalid Project Key",
			fmt.Sprintf("The provided project key '%s' is not valid. Project keys must be uppercase alphanumeric with underscores, 2-128 characters.", projectKey),
		)
		return
	}

	// Set the key attribute
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}

// projectKeyValidator validates project keys
type projectKeyValidator struct{}

func (v *projectKeyValidator) Description(ctx context.Context) string {
	return "Project key must be uppercase alphanumeric with underscores or hyphens, 2-128 characters"
}

func (v *projectKeyValidator) MarkdownDescription(ctx context.Context) string {
	return "Project key must be uppercase alphanumeric with underscores or hyphens, 2-128 characters"
}

func (v *projectKeyValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	key := req.ConfigValue.ValueString()

	if !isValidProjectKey(key) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Project Key",
			fmt.Sprintf("Project key '%s' is invalid. Must be uppercase alphanumeric with underscores or hyphens, 2-128 characters. "+
				"Example: MYPROJ, MY_PROJECT, ALPINA-SANDBOX", key),
		)
	}
}

// isValidProjectKey checks if a project key is valid
func isValidProjectKey(key string) bool {
	// Project key must be:
	// - 2 to 128 characters
	// - Uppercase letters, numbers, underscores, and hyphens
	// - Must start with an uppercase letter
	match, _ := regexp.MatchString(`^[A-Z][A-Z0-9_-]{1,127}$`, key)
	return match
}
