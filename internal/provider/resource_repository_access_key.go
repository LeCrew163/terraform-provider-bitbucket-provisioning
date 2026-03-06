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
	_ resource.Resource                = &repositoryAccessKeyResource{}
	_ resource.ResourceWithConfigure   = &repositoryAccessKeyResource{}
	_ resource.ResourceWithImportState = &repositoryAccessKeyResource{}
)

// NewRepositoryAccessKeyResource is a helper function to simplify the provider implementation.
func NewRepositoryAccessKeyResource() resource.Resource {
	return &repositoryAccessKeyResource{}
}

// repositoryAccessKeyResource is the resource implementation.
type repositoryAccessKeyResource struct {
	client *client.Client
}

// repositoryAccessKeyResourceModel maps the resource schema data.
type repositoryAccessKeyResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProjectKey     types.String `tfsdk:"project_key"`
	RepositorySlug types.String `tfsdk:"repository_slug"`
	KeyID          types.Int64  `tfsdk:"key_id"`
	PublicKey      types.String `tfsdk:"public_key"`
	Label          types.String `tfsdk:"label"`
	Permission     types.String `tfsdk:"permission"`
	Fingerprint    types.String `tfsdk:"fingerprint"`
}

// Metadata returns the resource type name.
func (r *repositoryAccessKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_access_key"
}

// Schema defines the schema for the resource.
func (r *repositoryAccessKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a single SSH access key for a Bitbucket Data Center repository. " +
			"Each resource represents one key granting read or write access to the repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier in the format {project_key}/{repository_slug}/{key_id}.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_key": schema.StringAttribute{
				Description: "The key of the project that owns the repository. Forces replacement if changed.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repository_slug": schema.StringAttribute{
				Description: "The slug of the repository to add the SSH access key to. Forces replacement if changed.",
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
				Description: "The SSH public key text (e.g. 'ssh-ed25519 AAAA...'). Forces replacement if changed.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"label": schema.StringAttribute{
				Description: "A human-readable label for the key. Bitbucket may derive a label from the SSH key " +
					"comment when none is supplied. Forces replacement if changed because Bitbucket does not " +
					"provide an update endpoint for labels.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"permission": schema.StringAttribute{
				Description: "The permission level granted to the key: REPO_READ or REPO_WRITE.",
				Required:    true,
				Validators: []validator.String{
					&repoAccessKeyPermissionValidator{},
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
func (r *repositoryAccessKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *repositoryAccessKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan repositoryAccessKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating repository access key", map[string]interface{}{
		"project_key":     plan.ProjectKey.ValueString(),
		"repository_slug": plan.RepositorySlug.ValueString(),
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
		AddForRepository(authCtx, plan.ProjectKey.ValueString(), plan.RepositorySlug.ValueString()).
		RestSshAccessKey(*sshAccessKey).Execute()

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
			resp.Diagnostics.AddError(
				"SSH Key Already Exists",
				fmt.Sprintf("An SSH key with the same text already exists for repository %q/%q. "+
					"Import or remove the existing key first.",
					plan.ProjectKey.ValueString(), plan.RepositorySlug.ValueString()),
			)
			return
		}
		// Bitbucket returns 2xx but the generated client fails to unmarshal the response
		// (DisallowUnknownFields rejects extra JSON fields in nested objects).
		// Recover by listing keys and finding the newly-created one by public key text.
		if httpResp != nil && httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
			tflog.Debug(ctx, "AddForRepository returned 2xx with unmarshal error; recovering via key text lookup", map[string]interface{}{
				"project_key":     plan.ProjectKey.ValueString(),
				"repository_slug": plan.RepositorySlug.ValueString(),
			})
			found, findErr := r.findKeyByText(ctx, plan.ProjectKey.ValueString(), plan.RepositorySlug.ValueString(), plan.PublicKey.ValueString())
			if findErr != nil {
				resp.Diagnostics.Append(client.HandleError("Failed to Create Repository Access Key", findErr)...)
				return
			}
			if found == nil {
				resp.Diagnostics.AddError("Failed to Create Repository Access Key", "Key was created but could not be retrieved from the server.")
				return
			}
			mapRepoAccessKeyToState(plan.ProjectKey.ValueString(), plan.RepositorySlug.ValueString(), found, &plan)
			plan.PublicKey = plannedPublicKey
			if !plannedLabel.IsNull() && !plannedLabel.IsUnknown() {
				plan.Label = plannedLabel
			}
			tflog.Debug(ctx, "Repository access key created (recovered)", map[string]interface{}{"id": plan.ID.ValueString()})
			resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
			return
		}
		resp.Diagnostics.Append(client.HandleError("Failed to Create Repository Access Key", client.ParseErrorResponse(httpResp))...)
		return
	}

	mapRepoAccessKeyToState(plan.ProjectKey.ValueString(), plan.RepositorySlug.ValueString(), created, &plan)
	plan.PublicKey = plannedPublicKey
	if !plannedLabel.IsNull() && !plannedLabel.IsUnknown() {
		plan.Label = plannedLabel
	}

	tflog.Debug(ctx, "Repository access key created", map[string]interface{}{
		"id": plan.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *repositoryAccessKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state repositoryAccessKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyID := state.KeyID.ValueInt64()
	tflog.Debug(ctx, "Reading repository access key", map[string]interface{}{
		"project_key":     state.ProjectKey.ValueString(),
		"repository_slug": state.RepositorySlug.ValueString(),
		"key_id":          keyID,
	})

	// Preserve values stored at creation to avoid drift from global key deduplication.
	savedPublicKey := state.PublicKey
	savedLabel := state.Label

	found, err := r.findKeyByID(ctx, state.ProjectKey.ValueString(), state.RepositorySlug.ValueString(), keyID)
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Read Repository Access Key", err)...)
		return
	}
	if found == nil {
		tflog.Warn(ctx, "Repository access key not found, removing from state", map[string]interface{}{
			"project_key":     state.ProjectKey.ValueString(),
			"repository_slug": state.RepositorySlug.ValueString(),
			"key_id":          keyID,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	mapRepoAccessKeyToState(state.ProjectKey.ValueString(), state.RepositorySlug.ValueString(), found, &state)
	if !savedPublicKey.IsNull() {
		state.PublicKey = savedPublicKey
	}
	if !savedLabel.IsNull() {
		state.Label = savedLabel
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update updates the permission level in-place.
func (r *repositoryAccessKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state repositoryAccessKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyID := state.KeyID.ValueInt64()
	keyIDStr := strconv.FormatInt(keyID, 10)

	tflog.Debug(ctx, "Updating repository access key permission", map[string]interface{}{
		"project_key":     plan.ProjectKey.ValueString(),
		"repository_slug": plan.RepositorySlug.ValueString(),
		"key_id":          keyID,
		"permission":      plan.Permission.ValueString(),
	})

	authCtx := r.client.NewAuthContext(ctx)
	updated, httpResp, err := r.client.GetAPIClient().AuthenticationAPI.
		UpdatePermission1(authCtx, plan.ProjectKey.ValueString(), keyIDStr, plan.Permission.ValueString(), plan.RepositorySlug.ValueString()).
		Execute()

	if err != nil {
		// Same 2xx+unmarshal workaround as Create.
		if httpResp != nil && httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
			found, findErr := r.findKeyByID(ctx, plan.ProjectKey.ValueString(), plan.RepositorySlug.ValueString(), keyID)
			if findErr != nil {
				resp.Diagnostics.Append(client.HandleError("Failed to Update Repository Access Key Permission", findErr)...)
				return
			}
			if found != nil {
				mapRepoAccessKeyToState(plan.ProjectKey.ValueString(), plan.RepositorySlug.ValueString(), found, &plan)
				// public_key and label cannot change via Update (they are RequiresReplace);
				// preserve the state values to avoid drift from global key deduplication.
				plan.PublicKey = state.PublicKey
				plan.Label = state.Label
				resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
				return
			}
		}
		resp.Diagnostics.Append(client.HandleError("Failed to Update Repository Access Key Permission", client.ParseErrorResponse(httpResp))...)
		return
	}

	mapRepoAccessKeyToState(plan.ProjectKey.ValueString(), plan.RepositorySlug.ValueString(), updated, &plan)
	// public_key and label cannot change via Update (they are RequiresReplace);
	// preserve the state values to avoid drift from global key deduplication.
	plan.PublicKey = state.PublicKey
	plan.Label = state.Label
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete revokes the SSH access key.
func (r *repositoryAccessKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state repositoryAccessKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyID := state.KeyID.ValueInt64()
	keyIDStr := strconv.FormatInt(keyID, 10)

	tflog.Debug(ctx, "Deleting repository access key", map[string]interface{}{
		"project_key":     state.ProjectKey.ValueString(),
		"repository_slug": state.RepositorySlug.ValueString(),
		"key_id":          keyID,
	})

	authCtx := r.client.NewAuthContext(ctx)
	httpResp, err := r.client.GetAPIClient().AuthenticationAPI.
		RevokeForRepository(authCtx, state.ProjectKey.ValueString(), keyIDStr, state.RepositorySlug.ValueString()).Execute()

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			tflog.Warn(ctx, "Repository access key already deleted", map[string]interface{}{"key_id": keyID})
			return
		}
		resp.Diagnostics.Append(client.HandleError("Failed to Delete Repository Access Key", client.ParseErrorResponse(httpResp))...)
		return
	}

	tflog.Debug(ctx, "Repository access key deleted", map[string]interface{}{"key_id": keyID})
}

// ImportState imports the resource by project_key/repository_slug/key_id.
func (r *repositoryAccessKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in the format {project_key}/{repository_slug}/{key_id}, got: %q", req.ID),
		)
		return
	}

	keyID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Key ID",
			fmt.Sprintf("The key_id portion of the import ID must be a number, got: %q", parts[2]),
		)
		return
	}

	found, findErr := r.findKeyByID(ctx, parts[0], parts[1], keyID)
	if findErr != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Import Repository Access Key", findErr)...)
		return
	}
	if found == nil {
		resp.Diagnostics.AddError(
			"Access Key Not Found",
			fmt.Sprintf("No SSH access key with id %d found in repository %q/%q.", keyID, parts[0], parts[1]),
		)
		return
	}

	var state repositoryAccessKeyResourceModel
	mapRepoAccessKeyToState(parts[0], parts[1], found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// findKeyByID searches the paginated key list for a key matching the given ID.
func (r *repositoryAccessKeyResource) findKeyByID(ctx context.Context, projectKey, repoSlug string, keyID int64) (*bitbucket.RestSshAccessKey, error) {
	authCtx := r.client.NewAuthContext(ctx)
	var start float32

	for {
		page, httpResp, err := r.client.GetAPIClient().AuthenticationAPI.
			GetForRepository1(authCtx, projectKey, repoSlug).
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
func (r *repositoryAccessKeyResource) findKeyByText(ctx context.Context, projectKey, repoSlug, publicKeyText string) (*bitbucket.RestSshAccessKey, error) {
	authCtx := r.client.NewAuthContext(ctx)
	var start float32

	for {
		page, httpResp, err := r.client.GetAPIClient().AuthenticationAPI.
			GetForRepository1(authCtx, projectKey, repoSlug).
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

// mapRepoAccessKeyToState populates a Terraform model from an API RestSshAccessKey response.
func mapRepoAccessKeyToState(projectKey, repoSlug string, key *bitbucket.RestSshAccessKey, state *repositoryAccessKeyResourceModel) {
	keyData, _ := key.GetKeyOk()

	var keyID int64
	if keyData != nil {
		keyID = int64(keyData.GetId())
	}

	state.ID = types.StringValue(fmt.Sprintf("%s/%s/%d", projectKey, repoSlug, keyID))
	state.ProjectKey = types.StringValue(projectKey)
	state.RepositorySlug = types.StringValue(repoSlug)
	state.KeyID = types.Int64Value(keyID)
	state.Permission = types.StringValue(key.GetPermission())

	if keyData != nil {
		state.PublicKey = types.StringValue(keyData.GetText())
		state.Fingerprint = types.StringValue(keyData.GetFingerprint())

		if label := keyData.GetLabel(); label != "" {
			state.Label = types.StringValue(label)
		} else {
			state.Label = types.StringNull()
		}
	}
}

// ── Validator ─────────────────────────────────────────────────────────────────

var validRepoAccessKeyPermissions = map[string]bool{
	"REPO_READ":  true,
	"REPO_WRITE": true,
}

type repoAccessKeyPermissionValidator struct{}

func (v *repoAccessKeyPermissionValidator) Description(_ context.Context) string {
	return "Access key permission must be one of: REPO_READ, REPO_WRITE"
}

func (v *repoAccessKeyPermissionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *repoAccessKeyPermissionValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !validRepoAccessKeyPermissions[req.ConfigValue.ValueString()] {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Repository Access Key Permission",
			fmt.Sprintf("%q is not a valid permission. Must be one of: REPO_READ, REPO_WRITE.", req.ConfigValue.ValueString()),
		)
	}
}
