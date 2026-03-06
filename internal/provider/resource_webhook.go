package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	bitbucket "github.com/LeCrew163/bitbucket-provisioning/internal/client/generated"
	"github.com/LeCrew163/bitbucket-provisioning/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &webhookResource{}
	_ resource.ResourceWithConfigure   = &webhookResource{}
	_ resource.ResourceWithImportState = &webhookResource{}
)

// NewWebhookResource is a helper function to simplify the provider implementation.
func NewWebhookResource() resource.Resource {
	return &webhookResource{}
}

type webhookResource struct {
	client *client.Client
}

type webhookResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	WebhookID               types.Int64  `tfsdk:"webhook_id"`
	ProjectKey              types.String `tfsdk:"project_key"`
	RepositorySlug          types.String `tfsdk:"repository_slug"`
	Name                    types.String `tfsdk:"name"`
	URL                     types.String `tfsdk:"url"`
	Events                  types.Set    `tfsdk:"events"`
	Active                  types.Bool   `tfsdk:"active"`
	SslVerificationRequired types.Bool   `tfsdk:"ssl_verification_required"`
}

func (r *webhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *webhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Bitbucket Data Center webhook for a project or repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier: {project_key}/{webhook_id} or {project_key}/{repository_slug}/{webhook_id}.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"webhook_id": schema.Int64Attribute{
				Description: "The numeric webhook ID assigned by Bitbucket.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
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
				Description: "The repository slug. When set the webhook is scoped to the repository; when omitted it is scoped to the project.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the webhook.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "The URL that Bitbucket will POST events to.",
				Required:    true,
			},
			"events": schema.SetAttribute{
				Description: "Event keys that trigger this webhook (e.g. repo:refs_changed, pr:opened).",
				Required:    true,
				ElementType: types.StringType,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the webhook is active. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"ssl_verification_required": schema.BoolAttribute{
				Description: "Whether SSL certificate verification is enforced. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

func (r *webhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *webhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data webhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.ProjectKey.ValueString()
	repoSlug := data.RepositorySlug.ValueString()

	var events []string
	resp.Diagnostics.Append(data.Events.ElementsAs(ctx, &events, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating webhook", map[string]interface{}{
		"project_key":     projectKey,
		"repository_slug": repoSlug,
		"name":            data.Name.ValueString(),
	})

	payload := buildWebhookPayload(&data, events)
	authCtx := r.client.NewAuthContext(ctx)

	webhook, httpResp, err := r.createWebhook(authCtx, projectKey, repoSlug, payload)
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Create Webhook", client.ParseErrorResponse(httpResp))...)
		return
	}
	if !webhook.HasId() {
		resp.Diagnostics.AddError("Webhook Create Failed", "Bitbucket did not return a webhook ID in the response.")
		return
	}

	mapWebhookToState(ctx, webhook, projectKey, repoSlug, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *webhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data webhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.ProjectKey.ValueString()
	repoSlug := data.RepositorySlug.ValueString()
	webhookIDStr := strconv.FormatInt(data.WebhookID.ValueInt64(), 10)

	tflog.Debug(ctx, "Reading webhook", map[string]interface{}{
		"project_key":     projectKey,
		"repository_slug": repoSlug,
		"webhook_id":      webhookIDStr,
	})

	authCtx := r.client.NewAuthContext(ctx)
	webhook, httpResp, err := r.getWebhook(authCtx, projectKey, repoSlug, webhookIDStr)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(client.HandleError("Failed to Read Webhook", client.ParseErrorResponse(httpResp))...)
		return
	}

	mapWebhookToState(ctx, webhook, projectKey, repoSlug, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *webhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state webhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := plan.ProjectKey.ValueString()
	repoSlug := plan.RepositorySlug.ValueString()
	webhookIDStr := strconv.FormatInt(state.WebhookID.ValueInt64(), 10)

	var events []string
	resp.Diagnostics.Append(plan.Events.ElementsAs(ctx, &events, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating webhook", map[string]interface{}{
		"project_key":     projectKey,
		"repository_slug": repoSlug,
		"webhook_id":      webhookIDStr,
	})

	payload := buildWebhookPayload(&plan, events)
	authCtx := r.client.NewAuthContext(ctx)

	webhook, httpResp, err := r.updateWebhook(authCtx, projectKey, repoSlug, webhookIDStr, payload)
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Update Webhook", client.ParseErrorResponse(httpResp))...)
		return
	}

	mapWebhookToState(ctx, webhook, projectKey, repoSlug, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data webhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.ProjectKey.ValueString()
	repoSlug := data.RepositorySlug.ValueString()
	webhookIDStr := strconv.FormatInt(data.WebhookID.ValueInt64(), 10)

	tflog.Debug(ctx, "Deleting webhook", map[string]interface{}{
		"project_key":     projectKey,
		"repository_slug": repoSlug,
		"webhook_id":      webhookIDStr,
	})

	authCtx := r.client.NewAuthContext(ctx)
	httpResp, err := r.deleteWebhook(authCtx, projectKey, repoSlug, webhookIDStr)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.Append(client.HandleError("Failed to Delete Webhook", client.ParseErrorResponse(httpResp))...)
	}
}

func (r *webhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import formats:
	//   project_key/webhook_id                     (project-scoped)
	//   project_key/repository_slug/webhook_id     (repo-scoped)
	//
	// Because repository slugs may contain slashes (rare), we split on the last
	// separator to get the numeric ID, then the first part is always project_key.
	lastSlash := strings.LastIndex(req.ID, "/")
	if lastSlash < 0 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected project_key/webhook_id or project_key/repository_slug/webhook_id, got %q", req.ID),
		)
		return
	}

	prefix := req.ID[:lastSlash]
	webhookIDPart := req.ID[lastSlash+1:]

	webhookID, err := strconv.ParseInt(webhookIDPart, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("The last path segment must be a numeric webhook ID, got %q", webhookIDPart),
		)
		return
	}

	// Determine project_key and optional repository_slug from prefix.
	firstSlash := strings.Index(prefix, "/")
	var projectKey, repoSlug string
	if firstSlash < 0 {
		// No slash in prefix → project-scoped: prefix is project_key.
		projectKey = prefix
	} else {
		// Slash in prefix → repo-scoped: project_key / repository_slug.
		projectKey = prefix[:firstSlash]
		repoSlug = prefix[firstSlash+1:]
	}

	if projectKey == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "project_key must not be empty")
		return
	}

	// Read the webhook from the API to populate all fields.
	authCtx := r.client.NewAuthContext(ctx)
	webhookIDStr := strconv.FormatInt(webhookID, 10)
	webhook, httpResp, apiErr := r.getWebhook(authCtx, projectKey, repoSlug, webhookIDStr)
	if apiErr != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Import Webhook", client.ParseErrorResponse(httpResp))...)
		return
	}
	if webhook == nil {
		resp.Diagnostics.AddError(
			"Webhook Not Found",
			fmt.Sprintf("No webhook with id %d found.", webhookID),
		)
		return
	}

	var state webhookResourceModel
	mapWebhookToState(ctx, webhook, projectKey, repoSlug, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ── API dispatch helpers ──────────────────────────────────────────────────────

func (r *webhookResource) createWebhook(
	authCtx context.Context,
	projectKey, repoSlug string,
	payload bitbucket.RestWebhook,
) (*bitbucket.RestWebhook, *http.Response, error) {
	if repoSlug != "" {
		return r.client.GetAPIClient().RepositoryAPI.
			CreateWebhook1(authCtx, projectKey, repoSlug).
			RestWebhook(payload).Execute()
	}
	return r.client.GetAPIClient().ProjectAPI.
		CreateWebhook(authCtx, projectKey).
		RestWebhook(payload).Execute()
}

func (r *webhookResource) getWebhook(
	authCtx context.Context,
	projectKey, repoSlug, webhookID string,
) (*bitbucket.RestWebhook, *http.Response, error) {
	if repoSlug != "" {
		return r.client.GetAPIClient().RepositoryAPI.
			GetWebhook1(authCtx, projectKey, webhookID, repoSlug).Execute()
	}
	return r.client.GetAPIClient().ProjectAPI.
		GetWebhook(authCtx, projectKey, webhookID).Execute()
}

func (r *webhookResource) updateWebhook(
	authCtx context.Context,
	projectKey, repoSlug, webhookID string,
	payload bitbucket.RestWebhook,
) (*bitbucket.RestWebhook, *http.Response, error) {
	if repoSlug != "" {
		return r.client.GetAPIClient().RepositoryAPI.
			UpdateWebhook1(authCtx, projectKey, webhookID, repoSlug).
			RestWebhook(payload).Execute()
	}
	return r.client.GetAPIClient().ProjectAPI.
		UpdateWebhook(authCtx, projectKey, webhookID).
		RestWebhook(payload).Execute()
}

func (r *webhookResource) deleteWebhook(
	authCtx context.Context,
	projectKey, repoSlug, webhookID string,
) (*http.Response, error) {
	if repoSlug != "" {
		return r.client.GetAPIClient().RepositoryAPI.
			DeleteWebhook1(authCtx, projectKey, webhookID, repoSlug).Execute()
	}
	return r.client.GetAPIClient().ProjectAPI.
		DeleteWebhook(authCtx, projectKey, webhookID).Execute()
}

// ── State mapping helpers ─────────────────────────────────────────────────────

func buildWebhookPayload(data *webhookResourceModel, events []string) bitbucket.RestWebhook {
	name := data.Name.ValueString()
	url := data.URL.ValueString()
	active := data.Active.ValueBool()
	sslVerif := data.SslVerificationRequired.ValueBool()
	return bitbucket.RestWebhook{
		Name:                    &name,
		Url:                     &url,
		Events:                  events,
		Active:                  &active,
		SslVerificationRequired: &sslVerif,
	}
}

func mapWebhookToState(ctx context.Context, webhook *bitbucket.RestWebhook, projectKey, repoSlug string, data *webhookResourceModel) {
	webhookID := webhook.GetId()

	if repoSlug != "" {
		data.ID = types.StringValue(fmt.Sprintf("%s/%s/%d", projectKey, repoSlug, webhookID))
		data.RepositorySlug = types.StringValue(repoSlug)
	} else {
		data.ID = types.StringValue(fmt.Sprintf("%s/%d", projectKey, webhookID))
		data.RepositorySlug = types.StringNull()
	}

	data.WebhookID = types.Int64Value(webhookID)
	data.ProjectKey = types.StringValue(projectKey)
	data.Name = types.StringValue(webhook.GetName())
	data.URL = types.StringValue(webhook.GetUrl())
	data.Active = types.BoolValue(webhook.GetActive())
	data.SslVerificationRequired = types.BoolValue(webhook.GetSslVerificationRequired())

	events := webhook.GetEvents()
	if len(events) == 0 {
		data.Events = types.SetValueMust(types.StringType, nil)
	} else {
		eventsSet, _ := types.SetValueFrom(ctx, types.StringType, events)
		data.Events = eventsSet
	}
}
