package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/LeCrew163/bitbucket-provisioning/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &repositoryHookResource{}
	_ resource.ResourceWithConfigure   = &repositoryHookResource{}
	_ resource.ResourceWithImportState = &repositoryHookResource{}
)

// NewRepositoryHookResource is a helper function to simplify the provider implementation.
func NewRepositoryHookResource() resource.Resource {
	return &repositoryHookResource{}
}

type repositoryHookResource struct {
	client *client.Client
}

type repositoryHookResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProjectKey     types.String `tfsdk:"project_key"`
	RepositorySlug types.String `tfsdk:"repository_slug"`
	HookKey        types.String `tfsdk:"hook_key"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	SettingsJSON   types.String `tfsdk:"settings_json"`
}

func (r *repositoryHookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_hook"
}

func (r *repositoryHookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Bitbucket Data Center repository hook. Works with any plugin that uses the Bitbucket hook framework (e.g. Webhook to Jenkins for Bitbucket). The settings_json attribute accepts arbitrary JSON matching the plugin's settings schema.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier: {project_key}/{repository_slug}/{hook_key}.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_key": schema.StringAttribute{
				Description: "The project key.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repository_slug": schema.StringAttribute{
				Description: "The repository slug.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hook_key": schema.StringAttribute{
				Description: "The fully-qualified hook key (e.g. com.nerdwin15.stash-stash-webhook-jenkins:postReceiveHook).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the hook is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"settings_json": schema.StringAttribute{
				Description: "Plugin-specific settings as a JSON string. Use jsonencode() to construct it. The exact fields depend on the plugin. Defaults to '{}'.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *repositoryHookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *repositoryHookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data repositoryHookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.ProjectKey.ValueString()
	repoSlug := data.RepositorySlug.ValueString()
	hookKey := data.HookKey.ValueString()

	tflog.Debug(ctx, "Creating repository hook", map[string]interface{}{
		"project_key":     projectKey,
		"repository_slug": repoSlug,
		"hook_key":        hookKey,
	})

	settingsJSON := data.SettingsJSON.ValueString()
	if settingsJSON == "" {
		settingsJSON = "{}"
	}

	// Validate and normalise the JSON first.
	normalised, err := normaliseJSON(settingsJSON)
	if err != nil {
		resp.Diagnostics.AddError("Invalid settings_json", fmt.Sprintf("settings_json is not valid JSON: %s", err))
		return
	}

	// Apply settings.
	if err := r.putSettings(ctx, projectKey, repoSlug, hookKey, normalised); err != nil {
		resp.Diagnostics.AddError("Failed to Set Repository Hook Settings", err.Error())
		return
	}

	// Enable or disable.
	authCtx := r.client.NewAuthContext(ctx)
	if data.Enabled.ValueBool() {
		if _, httpResp, apiErr := r.client.GetAPIClient().RepositoryAPI.
			EnableHook1(authCtx, projectKey, hookKey, repoSlug).Execute(); apiErr != nil {
			resp.Diagnostics.Append(client.HandleError("Failed to Enable Repository Hook", client.ParseErrorResponse(httpResp))...)
			return
		}
	} else {
		if _, httpResp, apiErr := r.client.GetAPIClient().RepositoryAPI.
			DisableHook1(authCtx, projectKey, hookKey, repoSlug).Execute(); apiErr != nil {
			resp.Diagnostics.Append(client.HandleError("Failed to Disable Repository Hook", client.ParseErrorResponse(httpResp))...)
			return
		}
	}

	data.SettingsJSON = types.StringValue(normalised)
	data.ID = types.StringValue(repositoryHookID(projectKey, repoSlug, hookKey))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryHookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data repositoryHookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.ProjectKey.ValueString()
	repoSlug := data.RepositorySlug.ValueString()
	hookKey := data.HookKey.ValueString()

	tflog.Debug(ctx, "Reading repository hook", map[string]interface{}{
		"project_key":     projectKey,
		"repository_slug": repoSlug,
		"hook_key":        hookKey,
	})

	authCtx := r.client.NewAuthContext(ctx)

	// Read enabled state.
	hookDetails, httpResp, err := r.client.GetAPIClient().RepositoryAPI.
		GetRepositoryHook1(authCtx, projectKey, hookKey, repoSlug).Execute()
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(client.HandleError("Failed to Read Repository Hook", client.ParseErrorResponse(httpResp))...)
		return
	}
	data.Enabled = types.BoolValue(hookDetails.GetEnabled())

	// Read settings.
	settingsJSON, err := r.getSettings(ctx, projectKey, repoSlug, hookKey)
	if err != nil {
		resp.Diagnostics.AddError("Failed to Read Repository Hook Settings", err.Error())
		return
	}
	data.SettingsJSON = types.StringValue(settingsJSON)
	data.ID = types.StringValue(repositoryHookID(projectKey, repoSlug, hookKey))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryHookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data repositoryHookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.ProjectKey.ValueString()
	repoSlug := data.RepositorySlug.ValueString()
	hookKey := data.HookKey.ValueString()

	settingsJSON := data.SettingsJSON.ValueString()
	if settingsJSON == "" {
		settingsJSON = "{}"
	}

	normalised, err := normaliseJSON(settingsJSON)
	if err != nil {
		resp.Diagnostics.AddError("Invalid settings_json", fmt.Sprintf("settings_json is not valid JSON: %s", err))
		return
	}

	if err := r.putSettings(ctx, projectKey, repoSlug, hookKey, normalised); err != nil {
		resp.Diagnostics.AddError("Failed to Update Repository Hook Settings", err.Error())
		return
	}

	authCtx := r.client.NewAuthContext(ctx)
	if data.Enabled.ValueBool() {
		if _, httpResp, apiErr := r.client.GetAPIClient().RepositoryAPI.
			EnableHook1(authCtx, projectKey, hookKey, repoSlug).Execute(); apiErr != nil {
			resp.Diagnostics.Append(client.HandleError("Failed to Enable Repository Hook", client.ParseErrorResponse(httpResp))...)
			return
		}
	} else {
		if _, httpResp, apiErr := r.client.GetAPIClient().RepositoryAPI.
			DisableHook1(authCtx, projectKey, hookKey, repoSlug).Execute(); apiErr != nil {
			resp.Diagnostics.Append(client.HandleError("Failed to Disable Repository Hook", client.ParseErrorResponse(httpResp))...)
			return
		}
	}

	data.SettingsJSON = types.StringValue(normalised)
	data.ID = types.StringValue(repositoryHookID(projectKey, repoSlug, hookKey))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryHookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data repositoryHookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.ProjectKey.ValueString()
	repoSlug := data.RepositorySlug.ValueString()
	hookKey := data.HookKey.ValueString()

	tflog.Debug(ctx, "Deleting repository hook", map[string]interface{}{
		"project_key":     projectKey,
		"repository_slug": repoSlug,
		"hook_key":        hookKey,
	})

	// Disable the hook and reset settings to empty.
	authCtx := r.client.NewAuthContext(ctx)
	if _, httpResp, apiErr := r.client.GetAPIClient().RepositoryAPI.
		DisableHook1(authCtx, projectKey, hookKey, repoSlug).Execute(); apiErr != nil {
		if httpResp == nil || httpResp.StatusCode != http.StatusNotFound {
			resp.Diagnostics.Append(client.HandleError("Failed to Disable Repository Hook", client.ParseErrorResponse(httpResp))...)
			return
		}
	}

	// Reset settings to empty object.
	if err := r.putSettings(ctx, projectKey, repoSlug, hookKey, "{}"); err != nil {
		resp.Diagnostics.AddError("Failed to Reset Repository Hook Settings", err.Error())
	}
}

func (r *repositoryHookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by project_key/repository_slug/hook_key
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected format: project_key/repository_slug/hook_key",
		)
		return
	}

	projectKey := parts[0]
	repoSlug := parts[1]
	hookKey := parts[2]

	authCtx := r.client.NewAuthContext(ctx)

	hookDetails, httpResp, err := r.client.GetAPIClient().RepositoryAPI.
		GetRepositoryHook1(authCtx, projectKey, hookKey, repoSlug).Execute()
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Import Repository Hook", client.ParseErrorResponse(httpResp))...)
		return
	}

	settingsJSON, err := r.getSettings(ctx, projectKey, repoSlug, hookKey)
	if err != nil {
		resp.Diagnostics.AddError("Failed to Import Repository Hook Settings", err.Error())
		return
	}

	state := repositoryHookResourceModel{
		ID:             types.StringValue(repositoryHookID(projectKey, repoSlug, hookKey)),
		ProjectKey:     types.StringValue(projectKey),
		RepositorySlug: types.StringValue(repoSlug),
		HookKey:        types.StringValue(hookKey),
		Enabled:        types.BoolValue(hookDetails.GetEnabled()),
		SettingsJSON:   types.StringValue(settingsJSON),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ── Raw JSON helpers ──────────────────────────────────────────────────────────

// getSettings fetches the hook settings as normalised JSON via a raw HTTP GET.
func (r *repositoryHookResource) getSettings(ctx context.Context, projectKey, repoSlug, hookKey string) (string, error) {
	urlPath := fmt.Sprintf("%s/rest/api/latest/projects/%s/repos/%s/settings/hooks/%s/settings",
		r.client.GetBaseURL(), projectKey, repoSlug, hookKey)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, urlPath, nil)
	if err != nil {
		return "", err
	}
	r.applyAuth(httpReq)
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := r.client.GetHTTPClient().Do(httpReq)
	if err != nil {
		return "", err
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", err
	}

	// 204 = hook has no settings configured yet
	if httpResp.StatusCode == http.StatusNoContent {
		return "{}", nil
	}

	if httpResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET hook settings returned %d: %s", httpResp.StatusCode, string(body))
	}

	// Normalise (compact, sorted keys).
	return normaliseJSON(string(body))
}

// putSettings sends hook settings as raw JSON via a raw HTTP PUT.
func (r *repositoryHookResource) putSettings(ctx context.Context, projectKey, repoSlug, hookKey, settingsJSON string) error {
	urlPath := fmt.Sprintf("%s/rest/api/latest/projects/%s/repos/%s/settings/hooks/%s/settings",
		r.client.GetBaseURL(), projectKey, repoSlug, hookKey)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, urlPath, bytes.NewBufferString(settingsJSON))
	if err != nil {
		return err
	}
	r.applyAuth(httpReq)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := r.client.GetHTTPClient().Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	body, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("PUT hook settings returned %d: %s", httpResp.StatusCode, string(body))
	}
	return nil
}

func (r *repositoryHookResource) applyAuth(req *http.Request) {
	if token := r.client.GetToken(); token != "" {
		req.SetBasicAuth(token, "")
	} else {
		req.SetBasicAuth(r.client.GetUsername(), r.client.GetPassword())
	}
}

// normaliseJSON unmarshals and re-marshals JSON to produce a compact, stable representation.
func normaliseJSON(s string) (string, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return "", err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func repositoryHookID(projectKey, repoSlug, hookKey string) string {
	return projectKey + "/" + repoSlug + "/" + hookKey
}
