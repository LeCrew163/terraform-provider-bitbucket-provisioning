package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/LeCrew163/bitbucket-provisioning/internal/client"
	bitbucket "github.com/LeCrew163/bitbucket-provisioning/internal/client/generated"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &projectAccessKeyResource{}
	_ resource.ResourceWithConfigure   = &projectAccessKeyResource{}
	_ resource.ResourceWithImportState = &projectAccessKeyResource{}
)

// NewProjectAccessKeyResource is a helper function to simplify the provider implementation.
func NewProjectAccessKeyResource() resource.Resource {
	return &projectAccessKeyResource{}
}

// projectAccessKeyResource is the resource implementation.
type projectAccessKeyResource struct {
	client *client.Client
}

// projectAccessKeyResourceModel maps the resource schema data.
type projectAccessKeyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectKey  types.String `tfsdk:"project_key"`
	KeyID       types.Int64  `tfsdk:"key_id"`
	PublicKey   types.String `tfsdk:"public_key"`
	Label       types.String `tfsdk:"label"`
	Permission  types.String `tfsdk:"permission"`
	Fingerprint types.String `tfsdk:"fingerprint"`
}

// Metadata returns the resource type name.
func (r *projectAccessKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_access_key"
}

// Schema defines the schema for the resource.
func (r *projectAccessKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a single SSH access key for a Bitbucket Data Center project. " +
			"Each resource represents one key granting read or write access to all repositories in the project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier in the format {project_key}/{key_id}.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_key": schema.StringAttribute{
				Description: "The key of the project to add the SSH access key to. Forces replacement if changed.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_id": schema.Int64Attribute{
				Description: "The server-assigned numeric identifier of the SSH key.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"public_key": schema.StringAttribute{
				Description: "The SSH public key text (e.g. 'ssh-rsa AAAA...'). Forces replacement if changed.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"label": schema.StringAttribute{
				Description: "A human-readable label for the key. Bitbucket may derive a label " +
					"from the SSH key comment when none is supplied. Forces replacement if changed " +
					"because Bitbucket does not provide an update endpoint for labels.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"permission": schema.StringAttribute{
				Description: "The permission level granted to the key: PROJECT_READ or PROJECT_WRITE.",
				Required:    true,
				Validators: []validator.String{
					&accessKeyPermissionValidator{},
				},
			},
			"fingerprint": schema.StringAttribute{
				Description: "The SSH key fingerprint (computed by Bitbucket).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *projectAccessKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *projectAccessKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectAccessKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating project access key", map[string]interface{}{
		"project_key": plan.ProjectKey.ValueString(),
	})

	// Bitbucket deduplicates SSH keys globally by key material, so the API may
	// return a pre-existing key object with different text/label metadata.
	// Preserve the planned values so state stays consistent with config.
	plannedPublicKey := plan.PublicKey
	plannedLabel := plan.Label

	keyData := bitbucket.NewAddSshKeyRequest()
	keyData.SetText(plan.PublicKey.ValueString())
	if !plan.Label.IsNull() && !plan.Label.IsUnknown() {
		keyData.SetLabel(plan.Label.ValueString())
	}

	sshAccessKey := bitbucket.NewRestSshAccessKey()
	sshAccessKey.SetKey(*keyData)
	sshAccessKey.SetPermission(plan.Permission.ValueString())

	authCtx := r.client.NewAuthContext(ctx)
	created, httpResp, err := r.client.GetAPIClient().AuthenticationAPI.
		AddForProject(authCtx, plan.ProjectKey.ValueString()).
		RestSshAccessKey(*sshAccessKey).Execute()

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
			resp.Diagnostics.AddError(
				"SSH Key Already Exists",
				fmt.Sprintf("An SSH key with the same text already exists for project %q. "+
					"Import or remove the existing key first.",
					plan.ProjectKey.ValueString()),
			)
			return
		}
		// Bitbucket returns 2xx but the generated client fails to unmarshal the
		// response because DisallowUnknownFields rejects extra JSON fields
		// (e.g. "links", "public") in the nested project object.  Treat any 2xx
		// as success and recover by searching for the newly-created key.
		if httpResp != nil && httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
			tflog.Debug(ctx, "AddForProject returned 2xx with unmarshal error; recovering via key text lookup", map[string]interface{}{
				"project_key": plan.ProjectKey.ValueString(),
			})
			found, findErr := r.findKeyByText(ctx, plan.ProjectKey.ValueString(), plan.PublicKey.ValueString())
			if findErr != nil {
				resp.Diagnostics.Append(client.HandleError("Failed to Create Project Access Key", findErr)...)
				return
			}
			if found == nil {
				resp.Diagnostics.AddError("Failed to Create Project Access Key", "Key was created but could not be retrieved from the server.")
				return
			}
			mapAccessKeyToState(plan.ProjectKey.ValueString(), found, &plan)
			plan.PublicKey = plannedPublicKey
			if !plannedLabel.IsNull() && !plannedLabel.IsUnknown() {
				plan.Label = plannedLabel
			}
			tflog.Debug(ctx, "Project access key created (recovered)", map[string]interface{}{"id": plan.ID.ValueString()})
			resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
			return
		}
		resp.Diagnostics.Append(client.HandleError("Failed to Create Project Access Key", client.ParseErrorResponse(httpResp))...)
		return
	}

	mapAccessKeyToState(plan.ProjectKey.ValueString(), created, &plan)
	plan.PublicKey = plannedPublicKey
	if !plannedLabel.IsNull() && !plannedLabel.IsUnknown() {
		plan.Label = plannedLabel
	}

	tflog.Debug(ctx, "Project access key created", map[string]interface{}{
		"id": plan.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *projectAccessKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectAccessKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyID := state.KeyID.ValueInt64()
	tflog.Debug(ctx, "Reading project access key", map[string]interface{}{
		"project_key": state.ProjectKey.ValueString(),
		"key_id":      keyID,
	})

	// Preserve values stored at creation to avoid drift from global key deduplication.
	savedPublicKey := state.PublicKey
	savedLabel := state.Label

	found, err := r.findKeyByID(ctx, state.ProjectKey.ValueString(), keyID)
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Read Project Access Key", err)...)
		return
	}
	if found == nil {
		tflog.Warn(ctx, "Project access key not found, removing from state", map[string]interface{}{
			"project_key": state.ProjectKey.ValueString(),
			"key_id":      keyID,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	mapAccessKeyToState(state.ProjectKey.ValueString(), found, &state)
	if !savedPublicKey.IsNull() {
		state.PublicKey = savedPublicKey
	}
	if !savedLabel.IsNull() {
		state.Label = savedLabel
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update updates the permission level in-place.
func (r *projectAccessKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state projectAccessKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyID := state.KeyID.ValueInt64()
	keyIDStr := strconv.FormatInt(keyID, 10)

	tflog.Debug(ctx, "Updating project access key permission", map[string]interface{}{
		"project_key": plan.ProjectKey.ValueString(),
		"key_id":      keyID,
		"permission":  plan.Permission.ValueString(),
	})

	authCtx := r.client.NewAuthContext(ctx)
	updated, httpResp, err := r.client.GetAPIClient().AuthenticationAPI.
		UpdatePermission(authCtx, plan.ProjectKey.ValueString(), keyIDStr, plan.Permission.ValueString()).
		Execute()

	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Update Project Access Key Permission", client.ParseErrorResponse(httpResp))...)
		return
	}

	mapAccessKeyToState(plan.ProjectKey.ValueString(), updated, &plan)
	// public_key and label cannot change via Update (they are RequiresReplace);
	// preserve the state values to avoid drift from global key deduplication.
	plan.PublicKey = state.PublicKey
	plan.Label = state.Label
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete revokes the SSH access key.
func (r *projectAccessKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectAccessKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyID := state.KeyID.ValueInt64()
	keyIDStr := strconv.FormatInt(keyID, 10)

	tflog.Debug(ctx, "Deleting project access key", map[string]interface{}{
		"project_key": state.ProjectKey.ValueString(),
		"key_id":      keyID,
	})

	authCtx := r.client.NewAuthContext(ctx)
	httpResp, err := r.client.GetAPIClient().AuthenticationAPI.
		RevokeForProject(authCtx, state.ProjectKey.ValueString(), keyIDStr).Execute()

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			tflog.Warn(ctx, "Project access key already deleted", map[string]interface{}{
				"key_id": keyID,
			})
			return
		}
		resp.Diagnostics.Append(client.HandleError("Failed to Delete Project Access Key", client.ParseErrorResponse(httpResp))...)
		return
	}

	tflog.Debug(ctx, "Project access key deleted", map[string]interface{}{
		"key_id": keyID,
	})
}

// ImportState imports the resource by project_key/key_id.
func (r *projectAccessKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in the format {project_key}/{key_id}, got: %q", req.ID),
		)
		return
	}

	keyID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Key ID",
			fmt.Sprintf("The key_id portion of the import ID must be a number, got: %q", parts[1]),
		)
		return
	}

	found, findErr := r.findKeyByID(ctx, parts[0], keyID)
	if findErr != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Import Project Access Key", findErr)...)
		return
	}
	if found == nil {
		resp.Diagnostics.AddError(
			"Access Key Not Found",
			fmt.Sprintf("No SSH access key with id %d found in project %q.", keyID, parts[0]),
		)
		return
	}

	var state projectAccessKeyResourceModel
	mapAccessKeyToState(parts[0], found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ── Helpers ───────────────────────────────────────────────────────────────

// findKeyByID searches the paginated key list for a key matching the given ID.
func (r *projectAccessKeyResource) findKeyByID(ctx context.Context, projectKey string, keyID int64) (*bitbucket.RestSshAccessKey, error) {
	authCtx := r.client.NewAuthContext(ctx)
	var start float32

	for {
		page, httpResp, err := r.client.GetAPIClient().AuthenticationAPI.
			GetSshKeysForProject(authCtx, projectKey).
			Start(start).Limit(100).Execute()
		if err != nil {
			return nil, client.ParseErrorResponse(httpResp)
		}

		for _, key := range page.GetValues() {
			if kd, ok := key.GetKeyOk(); ok && int64(kd.GetId()) == keyID {
				k := key
				return &k, nil
			}
		}

		if page.GetIsLastPage() {
			break
		}
		start = float32(page.GetNextPageStart())
	}

	return nil, nil
}

// findKeyByText searches the paginated key list for a key whose public-key text matches.
// Used to recover after Create gets a 2xx + unmarshal error from the generated client.
func (r *projectAccessKeyResource) findKeyByText(ctx context.Context, projectKey, publicKeyText string) (*bitbucket.RestSshAccessKey, error) {
	authCtx := r.client.NewAuthContext(ctx)
	var start float32

	for {
		page, httpResp, err := r.client.GetAPIClient().AuthenticationAPI.
			GetSshKeysForProject(authCtx, projectKey).
			Start(start).Limit(100).Execute()
		if err != nil {
			return nil, client.ParseErrorResponse(httpResp)
		}

		for _, key := range page.GetValues() {
			if kd, ok := key.GetKeyOk(); ok && normalizeSSHKey(kd.GetText()) == normalizeSSHKey(publicKeyText) {
				k := key
				return &k, nil
			}
		}

		if page.GetIsLastPage() {
			break
		}
		start = float32(page.GetNextPageStart())
	}

	return nil, nil
}

// mapAccessKeyToState populates a Terraform model from an API RestSshAccessKey response.
func mapAccessKeyToState(projectKey string, key *bitbucket.RestSshAccessKey, state *projectAccessKeyResourceModel) {
	keyData, _ := key.GetKeyOk()

	var keyID int64
	if keyData != nil {
		keyID = int64(keyData.GetId())
	}

	state.ID = types.StringValue(fmt.Sprintf("%s/%d", projectKey, keyID))
	state.ProjectKey = types.StringValue(projectKey)
	state.KeyID = types.Int64Value(keyID)
	state.Permission = types.StringValue(key.GetPermission())

	if keyData != nil {
		// Bitbucket normalizes the SSH key comment field (the third space-separated
		// component) and may return a different value than what was submitted. Store
		// only "algorithm base64" so the state stays stable across re-reads.
		state.PublicKey = types.StringValue(keyData.GetText())
		state.Fingerprint = types.StringValue(keyData.GetFingerprint())

		if label := keyData.GetLabel(); label != "" {
			state.Label = types.StringValue(label)
		} else {
			state.Label = types.StringNull()
		}
	}
}

// normalizeSSHKey strips the optional comment from an SSH public key, keeping
// only the "algorithm base64" portion. Bitbucket may normalize or replace the
// comment, so storing only the key material avoids spurious state drift.
func normalizeSSHKey(text string) string {
	parts := strings.Fields(text)
	if len(parts) >= 2 {
		return parts[0] + " " + parts[1]
	}
	return text
}

// ── Validator ─────────────────────────────────────────────────────────────

var validAccessKeyPermissions = map[string]bool{
	"PROJECT_READ":  true,
	"PROJECT_WRITE": true,
}

type accessKeyPermissionValidator struct{}

func (v *accessKeyPermissionValidator) Description(_ context.Context) string {
	return "Access key permission must be one of: PROJECT_READ, PROJECT_WRITE"
}

func (v *accessKeyPermissionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *accessKeyPermissionValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !validAccessKeyPermissions[req.ConfigValue.ValueString()] {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Access Key Permission",
			fmt.Sprintf("%q is not a valid permission. Must be one of: PROJECT_READ, PROJECT_WRITE.", req.ConfigValue.ValueString()),
		)
	}
}
