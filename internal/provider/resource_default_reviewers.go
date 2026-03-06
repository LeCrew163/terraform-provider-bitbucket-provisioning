package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	bitbucket "github.com/LeCrew163/bitbucket-provisioning/internal/client/generated"
	"github.com/LeCrew163/bitbucket-provisioning/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &defaultReviewersResource{}
	_ resource.ResourceWithConfigure   = &defaultReviewersResource{}
	_ resource.ResourceWithImportState = &defaultReviewersResource{}
)

// NewDefaultReviewersResource is a helper function to simplify the provider implementation.
func NewDefaultReviewersResource() resource.Resource {
	return &defaultReviewersResource{}
}

type defaultReviewersResource struct {
	client *client.Client
}

// conditionModel represents one default-reviewer condition block.
// No computed ID is stored: conditions are identified by their source+target
// matcher combination (semantic key) for all reconciliation operations.
type conditionModel struct {
	SourceMatcherType types.String `tfsdk:"source_matcher_type"`
	SourceMatcherID   types.String `tfsdk:"source_matcher_id"`
	TargetMatcherType types.String `tfsdk:"target_matcher_type"`
	TargetMatcherID   types.String `tfsdk:"target_matcher_id"`
	Users             types.List   `tfsdk:"users"`
	RequiredApprovals types.Int64  `tfsdk:"required_approvals"`
}

type defaultReviewersResourceModel struct {
	ID             types.String     `tfsdk:"id"`
	ProjectKey     types.String     `tfsdk:"project_key"`
	RepositorySlug types.String     `tfsdk:"repository_slug"`
	Conditions     []conditionModel `tfsdk:"condition"`
}

func (r *defaultReviewersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_default_reviewers"
}

func (r *defaultReviewersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Bitbucket Data Center default reviewer conditions for a project or repository. Reconciles the full set of conditions declared in the resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier: {project_key} or {project_key}/{repository_slug}.",
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
				Description: "The repository slug. When set, conditions are scoped to the repository; when omitted they are scoped to the project.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"condition": schema.SetNestedBlock{
				Description: "A default reviewer condition. Each block defines a rule that automatically adds reviewers to pull requests matching the source and target branch criteria. Conditions are identified by their source+target matcher combination.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"source_matcher_type": schema.StringAttribute{
							Description: "Type of the source branch matcher: ANY_REF, BRANCH, PATTERN, MODEL_BRANCH, or MODEL_CATEGORY.",
							Required:    true,
						},
						"source_matcher_id": schema.StringAttribute{
							Description: "The source branch matcher value (e.g. ANY_REF_MATCHER_ID, refs/heads/main, feature/*).",
							Required:    true,
						},
						"target_matcher_type": schema.StringAttribute{
							Description: "Type of the target branch matcher: ANY_REF, BRANCH, PATTERN, MODEL_BRANCH, or MODEL_CATEGORY.",
							Required:    true,
						},
						"target_matcher_id": schema.StringAttribute{
							Description: "The target branch matcher value (e.g. refs/heads/main, release/*).",
							Required:    true,
						},
						"users": schema.ListAttribute{
							Description: "List of user slugs (usernames) to add as required reviewers.",
							Required:    true,
							ElementType: types.StringType,
						},
						"required_approvals": schema.Int64Attribute{
							Description: "Minimum number of approvals required from the listed reviewers before the PR can be merged.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *defaultReviewersResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *defaultReviewersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data defaultReviewersResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.ProjectKey.ValueString()
	repoSlug := data.RepositorySlug.ValueString()

	tflog.Debug(ctx, "Creating default reviewers", map[string]interface{}{
		"project_key":     projectKey,
		"repository_slug": repoSlug,
	})

	authCtx := r.client.NewAuthContext(ctx)
	for i := range data.Conditions {
		cond := &data.Conditions[i]
		if diags := r.applyConditionCreate(authCtx, ctx, projectKey, repoSlug, cond); diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	setDefaultReviewersID(&data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *defaultReviewersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data defaultReviewersResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.ProjectKey.ValueString()
	repoSlug := data.RepositorySlug.ValueString()

	tflog.Debug(ctx, "Reading default reviewers", map[string]interface{}{
		"project_key":     projectKey,
		"repository_slug": repoSlug,
	})

	authCtx := r.client.NewAuthContext(ctx)
	conditions, httpResp, err := r.listConditions(authCtx, projectKey, repoSlug)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(client.HandleError("Failed to Read Default Reviewers", client.ParseErrorResponse(httpResp))...)
		return
	}

	// Build a lookup map by semantic key → API condition.
	apiByKey := make(map[string]bitbucket.RestPullRequestCondition, len(conditions))
	for _, c := range conditions {
		apiByKey[apiConditionKey(c)] = c
	}

	// Reconcile state: update each stored condition from API if it still exists.
	surviving := data.Conditions[:0]
	for _, stored := range data.Conditions {
		key := conditionKey(stored)
		apiCond, exists := apiByKey[key]
		if !exists {
			continue
		}
		updated, diags := mapConditionToModel(ctx, apiCond)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		surviving = append(surviving, updated)
	}
	data.Conditions = surviving

	setDefaultReviewersID(&data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *defaultReviewersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan defaultReviewersResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state defaultReviewersResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := plan.ProjectKey.ValueString()
	repoSlug := plan.RepositorySlug.ValueString()

	tflog.Debug(ctx, "Updating default reviewers", map[string]interface{}{
		"project_key":     projectKey,
		"repository_slug": repoSlug,
	})

	authCtx := r.client.NewAuthContext(ctx)

	// List current conditions from the API to get IDs for update/delete.
	current, httpResp, err := r.listConditions(authCtx, projectKey, repoSlug)
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to List Default Reviewer Conditions", client.ParseErrorResponse(httpResp))...)
		return
	}

	// Build lookup maps.
	apiByKey := make(map[string]bitbucket.RestPullRequestCondition, len(current))
	for _, c := range current {
		apiByKey[apiConditionKey(c)] = c
	}

	planKeys := make(map[string]struct{}, len(plan.Conditions))
	for _, c := range plan.Conditions {
		planKeys[conditionKey(c)] = struct{}{}
	}

	// Create or update conditions from the plan.
	for i := range plan.Conditions {
		desired := &plan.Conditions[i]
		key := conditionKey(*desired)

		if existing, found := apiByKey[key]; found {
			// Exists — update in place.
			condIDStr := strconv.FormatInt(int64(existing.GetId()), 10)
			if diags := r.applyConditionUpdate(authCtx, ctx, projectKey, repoSlug, condIDStr, desired); diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
		} else {
			// New — create it.
			if diags := r.applyConditionCreate(authCtx, ctx, projectKey, repoSlug, desired); diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
		}
	}

	// Delete conditions no longer in the plan.
	for _, apiCond := range current {
		key := apiConditionKey(apiCond)
		if _, keep := planKeys[key]; keep {
			continue
		}
		condIDStr := strconv.FormatInt(int64(apiCond.GetId()), 10)
		delResp, delErr := r.deleteCondition(authCtx, projectKey, repoSlug, condIDStr)
		if delErr != nil {
			if delResp != nil && delResp.StatusCode == http.StatusNotFound {
				continue
			}
			resp.Diagnostics.Append(client.HandleError("Failed to Delete Default Reviewer Condition", client.ParseErrorResponse(delResp))...)
			return
		}
	}

	setDefaultReviewersID(&plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *defaultReviewersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data defaultReviewersResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectKey := data.ProjectKey.ValueString()
	repoSlug := data.RepositorySlug.ValueString()

	tflog.Debug(ctx, "Deleting default reviewers", map[string]interface{}{
		"project_key":     projectKey,
		"repository_slug": repoSlug,
	})

	authCtx := r.client.NewAuthContext(ctx)

	// List to get actual IDs.
	conditions, httpResp, err := r.listConditions(authCtx, projectKey, repoSlug)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.Append(client.HandleError("Failed to List Default Reviewer Conditions for Delete", client.ParseErrorResponse(httpResp))...)
		return
	}

	for _, cond := range conditions {
		condIDStr := strconv.FormatInt(int64(cond.GetId()), 10)
		delResp, delErr := r.deleteCondition(authCtx, projectKey, repoSlug, condIDStr)
		if delErr != nil {
			if delResp != nil && delResp.StatusCode == http.StatusNotFound {
				continue
			}
			resp.Diagnostics.Append(client.HandleError("Failed to Delete Default Reviewer Condition", client.ParseErrorResponse(delResp))...)
			return
		}
	}
}

func (r *defaultReviewersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by project_key or project_key/repository_slug.
	parts := strings.SplitN(req.ID, "/", 2)
	projectKey := parts[0]
	repoSlug := ""
	if len(parts) == 2 {
		repoSlug = parts[1]
	}

	if projectKey == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "project_key must not be empty")
		return
	}

	authCtx := r.client.NewAuthContext(ctx)
	conditions, httpResp, err := r.listConditions(authCtx, projectKey, repoSlug)
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Import Default Reviewers", client.ParseErrorResponse(httpResp))...)
		return
	}

	var state defaultReviewersResourceModel
	state.ProjectKey = types.StringValue(projectKey)
	if repoSlug != "" {
		state.RepositorySlug = types.StringValue(repoSlug)
	} else {
		state.RepositorySlug = types.StringNull()
	}

	for _, c := range conditions {
		model, diags := mapConditionToModel(ctx, c)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Conditions = append(state.Conditions, model)
	}

	setDefaultReviewersID(&state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ── API dispatch helpers ──────────────────────────────────────────────────────

func (r *defaultReviewersResource) applyConditionCreate(
	authCtx, ctx context.Context,
	projectKey, repoSlug string,
	cond *conditionModel,
) diag.Diagnostics {
	var slugs []string
	if d := cond.Users.ElementsAs(ctx, &slugs, false); d.HasError() {
		return d
	}

	resolvedUsers, err := r.resolveUsers(authCtx, slugs)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Failed to Resolve Reviewer Users", err.Error())
		return diags
	}

	payload := buildConditionPayload(cond, resolvedUsers)
	_, httpResp, apiErr := r.createCondition(authCtx, projectKey, repoSlug, payload)
	if apiErr != nil {
		return client.HandleError("Failed to Create Default Reviewer Condition", client.ParseErrorResponse(httpResp))
	}
	return nil
}

func (r *defaultReviewersResource) applyConditionUpdate(
	authCtx, ctx context.Context,
	projectKey, repoSlug, condIDStr string,
	cond *conditionModel,
) diag.Diagnostics {
	var slugs []string
	if d := cond.Users.ElementsAs(ctx, &slugs, false); d.HasError() {
		return d
	}

	resolvedUsers, err := r.resolveUsers(authCtx, slugs)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Failed to Resolve Reviewer Users", err.Error())
		return diags
	}

	payload := buildConditionPayload(cond, resolvedUsers)
	_, httpResp, apiErr := r.updateCondition(authCtx, projectKey, repoSlug, condIDStr, payload)
	if apiErr != nil {
		return client.HandleError("Failed to Update Default Reviewer Condition", client.ParseErrorResponse(httpResp))
	}
	return nil
}

func (r *defaultReviewersResource) listConditions(
	authCtx context.Context,
	projectKey, repoSlug string,
) ([]bitbucket.RestPullRequestCondition, *http.Response, error) {
	if repoSlug != "" {
		return r.client.GetAPIClient().PullRequestsAPI.
			GetPullRequestConditions1(authCtx, projectKey, repoSlug).Execute()
	}
	return r.client.GetAPIClient().PullRequestsAPI.
		GetPullRequestConditions(authCtx, projectKey).Execute()
}

func (r *defaultReviewersResource) createCondition(
	authCtx context.Context,
	projectKey, repoSlug string,
	payload bitbucket.RestDefaultReviewersRequest,
) (*bitbucket.RestPullRequestCondition, *http.Response, error) {
	if repoSlug != "" {
		return r.client.GetAPIClient().PullRequestsAPI.
			CreatePullRequestCondition1(authCtx, projectKey, repoSlug).
			RestDefaultReviewersRequest(payload).Execute()
	}
	return r.client.GetAPIClient().PullRequestsAPI.
		CreatePullRequestCondition(authCtx, projectKey).
		RestDefaultReviewersRequest(payload).Execute()
}

func (r *defaultReviewersResource) updateCondition(
	authCtx context.Context,
	projectKey, repoSlug, condID string,
	payload bitbucket.RestDefaultReviewersRequest,
) (*bitbucket.RestPullRequestCondition, *http.Response, error) {
	if repoSlug != "" {
		req := bitbucket.UpdatePullRequestCondition1Request{
			RequiredApprovals: payload.RequiredApprovals,
			Reviewers:         payload.Reviewers,
			SourceMatcher:     payload.SourceMatcher,
			TargetMatcher:     payload.TargetMatcher,
		}
		return r.client.GetAPIClient().PullRequestsAPI.
			UpdatePullRequestCondition1(authCtx, projectKey, condID, repoSlug).
			UpdatePullRequestCondition1Request(req).Execute()
	}
	return r.client.GetAPIClient().PullRequestsAPI.
		UpdatePullRequestCondition(authCtx, projectKey, condID).
		RestDefaultReviewersRequest(payload).Execute()
}

func (r *defaultReviewersResource) deleteCondition(
	authCtx context.Context,
	projectKey, repoSlug, condID string,
) (*http.Response, error) {
	if repoSlug != "" {
		condIDInt, _ := strconv.ParseInt(condID, 10, 32)
		return r.client.GetAPIClient().PullRequestsAPI.
			DeletePullRequestCondition1(authCtx, projectKey, int32(condIDInt), repoSlug).Execute()
	}
	return r.client.GetAPIClient().PullRequestsAPI.
		DeletePullRequestCondition(authCtx, projectKey, condID).Execute()
}

// resolveUsers looks up each slug via the SystemMaintenance API to get the full
// user object (including numeric ID) required by the default reviewer condition API.
func (r *defaultReviewersResource) resolveUsers(
	authCtx context.Context,
	slugs []string,
) ([]bitbucket.RestApplicationUser, error) {
	users := make([]bitbucket.RestApplicationUser, 0, len(slugs))
	for _, slug := range slugs {
		user, httpResp, err := r.client.GetAPIClient().SystemMaintenanceAPI.GetUser(authCtx, slug).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve user %q: %w", slug, client.ParseErrorResponse(httpResp))
		}
		users = append(users, *user)
	}
	return users, nil
}

// ── Payload / state-mapping helpers ──────────────────────────────────────────

func buildConditionPayload(cond *conditionModel, resolvedUsers []bitbucket.RestApplicationUser) bitbucket.RestDefaultReviewersRequest {
	reqApprovals := int32(cond.RequiredApprovals.ValueInt64())
	sourceMatcher := buildMatcher(cond.SourceMatcherType.ValueString(), cond.SourceMatcherID.ValueString())
	targetMatcher := buildMatcher(cond.TargetMatcherType.ValueString(), cond.TargetMatcherID.ValueString())
	return bitbucket.RestDefaultReviewersRequest{
		RequiredApprovals: &reqApprovals,
		Reviewers:         resolvedUsers,
		SourceMatcher:     &sourceMatcher,
		TargetMatcher:     &targetMatcher,
	}
}

func buildMatcher(matcherType, matcherID string) bitbucket.UpdatePullRequestCondition1RequestSourceMatcher {
	displayID := matcherID
	if matcherType == "ANY_REF" {
		displayID = "Any branch"
	}
	typeName := matcherTypeDisplayName(matcherType)
	mt := bitbucket.NewUpdatePullRequestCondition1RequestSourceMatcherType(matcherType, typeName)
	return bitbucket.UpdatePullRequestCondition1RequestSourceMatcher{
		Id:        &matcherID,
		DisplayId: &displayID,
		Type:      mt,
	}
}

func matcherTypeDisplayName(t string) string {
	switch t {
	case "ANY_REF":
		return "Any branch"
	case "BRANCH":
		return "Branch"
	case "PATTERN":
		return "Pattern"
	case "MODEL_BRANCH":
		return "Model branch"
	case "MODEL_CATEGORY":
		return "Model category"
	default:
		return t
	}
}

// mapConditionToModel converts an API condition to the Terraform model.
func mapConditionToModel(ctx context.Context, c bitbucket.RestPullRequestCondition) (conditionModel, diag.Diagnostics) {
	m := conditionModel{
		RequiredApprovals: types.Int64Value(int64(c.GetRequiredApprovals())),
	}

	if src := c.SourceRefMatcher; src != nil {
		m.SourceMatcherID = types.StringValue(src.GetId())
		if t := src.Type; t != nil {
			m.SourceMatcherType = types.StringValue(t.GetId())
		}
	}
	if tgt := c.TargetRefMatcher; tgt != nil {
		m.TargetMatcherID = types.StringValue(tgt.GetId())
		if t := tgt.Type; t != nil {
			m.TargetMatcherType = types.StringValue(t.GetId())
		}
	}

	// Extract user slugs from the Reviewers field.
	// The generated client maps reviewers as []RestReviewerGroup even though the API
	// actually returns individual ApplicationUser objects. Due to this type mismatch,
	// the user's login name is deserialized into RestReviewerGroup.Name.
	var userSlugs []string
	for _, rg := range c.GetReviewers() {
		name := rg.GetName()
		if name != "" {
			userSlugs = append(userSlugs, name)
		}
	}

	userAttrs := make([]attr.Value, len(userSlugs))
	for i, s := range userSlugs {
		userAttrs[i] = types.StringValue(s)
	}
	if len(userAttrs) == 0 {
		m.Users = types.ListValueMust(types.StringType, nil)
	} else {
		m.Users = types.ListValueMust(types.StringType, userAttrs)
	}

	return m, nil
}

// conditionKey returns a string key that uniquely identifies a condition by its
// source+target matcher combination (used for reconciliation).
func conditionKey(c conditionModel) string {
	return strings.Join([]string{
		c.SourceMatcherType.ValueString(),
		c.SourceMatcherID.ValueString(),
		c.TargetMatcherType.ValueString(),
		c.TargetMatcherID.ValueString(),
	}, "|")
}

// apiConditionKey returns a string key for an API condition.
func apiConditionKey(c bitbucket.RestPullRequestCondition) string {
	parts := make([]string, 4)
	if src := c.SourceRefMatcher; src != nil {
		if t := src.Type; t != nil {
			parts[0] = t.GetId()
		}
		parts[1] = src.GetId()
	}
	if tgt := c.TargetRefMatcher; tgt != nil {
		if t := tgt.Type; t != nil {
			parts[2] = t.GetId()
		}
		parts[3] = tgt.GetId()
	}
	return strings.Join(parts, "|")
}

func setDefaultReviewersID(data *defaultReviewersResourceModel) {
	if data.RepositorySlug.ValueString() != "" {
		data.ID = types.StringValue(data.ProjectKey.ValueString() + "/" + data.RepositorySlug.ValueString())
	} else {
		data.ID = types.StringValue(data.ProjectKey.ValueString())
	}
}
