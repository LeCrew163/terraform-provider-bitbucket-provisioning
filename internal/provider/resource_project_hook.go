package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/internal/client"
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
	_ resource.Resource                = &projectHookResource{}
	_ resource.ResourceWithConfigure   = &projectHookResource{}
	_ resource.ResourceWithImportState = &projectHookResource{}
)

// NewProjectHookResource is a helper function to simplify the provider implementation.
func NewProjectHookResource() resource.Resource {
	return &projectHookResource{}
}

type projectHookResource struct {
	client *client.Client
}

type projectHookResourceModel struct {
	ID           types.String `tfsdk:"id"`
	ProjectKey   types.String `tfsdk:"project_key"`
	HookKey      types.String `tfsdk:"hook_key"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	SettingsJSON types.String `tfsdk:"settings_json"`
}

func (r *projectHookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_hook"
}

func (r *projectHookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Bitbucket Data Center project-level hook. Project hooks apply to all repositories within the project. Works with any plugin that uses the Bitbucket hook framework (e.g. Webhook to Jenkins for Bitbucket). The settings_json attribute accepts arbitrary JSON matching the plugin's settings schema.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier: {project_key}/{hook_key}.",
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

func (r *projectHookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectHookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data projectHookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.ProjectKey.ValueString()
	hookKey := data.HookKey.ValueString()

	tflog.Debug(ctx, "Creating project hook", map[string]interface{}{
		"project_key": projectKey,
		"hook_key":    hookKey,
	})

	settingsJSON := data.SettingsJSON.ValueString()
	if settingsJSON == "" {
		settingsJSON = "{}"
	}

	normalised, err := normaliseJSON(settingsJSON)
	if err != nil {
		resp.Diagnostics.AddError("Invalid settings_json", fmt.Sprintf("settings_json is not valid JSON: %s", err))
		return
	}

	if err := r.putSettings(ctx, projectKey, hookKey, normalised); err != nil {
		resp.Diagnostics.AddError("Failed to Set Project Hook Settings", err.Error())
		return
	}

	authCtx := r.client.NewAuthContext(ctx)
	if data.Enabled.ValueBool() {
		if _, httpResp, apiErr := r.client.GetAPIClient().ProjectAPI.
			EnableHook(authCtx, projectKey, hookKey).Execute(); apiErr != nil {
			resp.Diagnostics.Append(client.HandleError("Failed to Enable Project Hook", client.ParseErrorResponse(httpResp))...)
			return
		}
	} else {
		if _, httpResp, apiErr := r.client.GetAPIClient().ProjectAPI.
			DisableHook(authCtx, projectKey, hookKey).Execute(); apiErr != nil {
			resp.Diagnostics.Append(client.HandleError("Failed to Disable Project Hook", client.ParseErrorResponse(httpResp))...)
			return
		}
	}

	data.SettingsJSON = types.StringValue(normalised)
	data.ID = types.StringValue(projectHookID(projectKey, hookKey))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *projectHookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data projectHookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.ProjectKey.ValueString()
	hookKey := data.HookKey.ValueString()

	tflog.Debug(ctx, "Reading project hook", map[string]interface{}{
		"project_key": projectKey,
		"hook_key":    hookKey,
	})

	authCtx := r.client.NewAuthContext(ctx)

	hookDetails, httpResp, err := r.client.GetAPIClient().ProjectAPI.
		GetRepositoryHook(authCtx, projectKey, hookKey).Execute()
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(client.HandleError("Failed to Read Project Hook", client.ParseErrorResponse(httpResp))...)
		return
	}
	data.Enabled = types.BoolValue(hookDetails.GetEnabled())

	settingsJSON, err := r.getSettings(ctx, projectKey, hookKey)
	if err != nil {
		resp.Diagnostics.AddError("Failed to Read Project Hook Settings", err.Error())
		return
	}
	data.SettingsJSON = types.StringValue(settingsJSON)
	data.ID = types.StringValue(projectHookID(projectKey, hookKey))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *projectHookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data projectHookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.ProjectKey.ValueString()
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

	if err := r.putSettings(ctx, projectKey, hookKey, normalised); err != nil {
		resp.Diagnostics.AddError("Failed to Update Project Hook Settings", err.Error())
		return
	}

	authCtx := r.client.NewAuthContext(ctx)
	if data.Enabled.ValueBool() {
		if _, httpResp, apiErr := r.client.GetAPIClient().ProjectAPI.
			EnableHook(authCtx, projectKey, hookKey).Execute(); apiErr != nil {
			resp.Diagnostics.Append(client.HandleError("Failed to Enable Project Hook", client.ParseErrorResponse(httpResp))...)
			return
		}
	} else {
		if _, httpResp, apiErr := r.client.GetAPIClient().ProjectAPI.
			DisableHook(authCtx, projectKey, hookKey).Execute(); apiErr != nil {
			resp.Diagnostics.Append(client.HandleError("Failed to Disable Project Hook", client.ParseErrorResponse(httpResp))...)
			return
		}
	}

	data.SettingsJSON = types.StringValue(normalised)
	data.ID = types.StringValue(projectHookID(projectKey, hookKey))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *projectHookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data projectHookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.ProjectKey.ValueString()
	hookKey := data.HookKey.ValueString()

	tflog.Debug(ctx, "Deleting project hook", map[string]interface{}{
		"project_key": projectKey,
		"hook_key":    hookKey,
	})

	authCtx := r.client.NewAuthContext(ctx)
	if _, httpResp, apiErr := r.client.GetAPIClient().ProjectAPI.
		DisableHook(authCtx, projectKey, hookKey).Execute(); apiErr != nil {
		if httpResp == nil || httpResp.StatusCode != http.StatusNotFound {
			resp.Diagnostics.Append(client.HandleError("Failed to Disable Project Hook", client.ParseErrorResponse(httpResp))...)
			return
		}
	}

	if err := r.putSettings(ctx, projectKey, hookKey, "{}"); err != nil {
		resp.Diagnostics.AddError("Failed to Reset Project Hook Settings", err.Error())
	}
}

func (r *projectHookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by project_key/hook_key
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected format: project_key/hook_key",
		)
		return
	}

	projectKey := parts[0]
	hookKey := parts[1]

	authCtx := r.client.NewAuthContext(ctx)

	hookDetails, httpResp, err := r.client.GetAPIClient().ProjectAPI.
		GetRepositoryHook(authCtx, projectKey, hookKey).Execute()
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Import Project Hook", client.ParseErrorResponse(httpResp))...)
		return
	}

	settingsJSON, err := r.getSettings(ctx, projectKey, hookKey)
	if err != nil {
		resp.Diagnostics.AddError("Failed to Import Project Hook Settings", err.Error())
		return
	}

	state := projectHookResourceModel{
		ID:           types.StringValue(projectHookID(projectKey, hookKey)),
		ProjectKey:   types.StringValue(projectKey),
		HookKey:      types.StringValue(hookKey),
		Enabled:      types.BoolValue(hookDetails.GetEnabled()),
		SettingsJSON: types.StringValue(settingsJSON),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ── Raw JSON helpers ──────────────────────────────────────────────────────────

func (r *projectHookResource) getSettings(ctx context.Context, projectKey, hookKey string) (string, error) {
	urlPath := fmt.Sprintf("%s/rest/api/latest/projects/%s/settings/hooks/%s/settings",
		r.client.GetBaseURL(), projectKey, hookKey)

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

	if httpResp.StatusCode == http.StatusNoContent {
		return "{}", nil
	}

	if httpResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET project hook settings returned %d: %s", httpResp.StatusCode, string(body))
	}

	return normaliseJSON(string(body))
}

func (r *projectHookResource) putSettings(ctx context.Context, projectKey, hookKey, settingsJSON string) error {
	urlPath := fmt.Sprintf("%s/rest/api/latest/projects/%s/settings/hooks/%s/settings",
		r.client.GetBaseURL(), projectKey, hookKey)

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
		return fmt.Errorf("PUT project hook settings returned %d: %s", httpResp.StatusCode, string(body))
	}
	return nil
}

func (r *projectHookResource) applyAuth(req *http.Request) {
	if token := r.client.GetToken(); token != "" {
		req.SetBasicAuth(token, "")
	} else {
		req.SetBasicAuth(r.client.GetUsername(), r.client.GetPassword())
	}
}

func projectHookID(projectKey, hookKey string) string {
	return projectKey + "/" + hookKey
}
