package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/LeCrew163/bitbucket-provisioning/internal/client"
	bitbucket "github.com/LeCrew163/bitbucket-provisioning/internal/client/generated"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	_ resource.Resource                = &branchPermissionsResource{}
	_ resource.ResourceWithConfigure   = &branchPermissionsResource{}
	_ resource.ResourceWithImportState = &branchPermissionsResource{}
)

// NewBranchPermissionsResource is a helper function to simplify the provider implementation.
func NewBranchPermissionsResource() resource.Resource {
	return &branchPermissionsResource{}
}

// branchPermissionsResource is the resource implementation.
type branchPermissionsResource struct {
	client *client.Client
}

// branchPermissionsResourceModel maps the resource schema data.
type branchPermissionsResourceModel struct {
	ID         types.String             `tfsdk:"id"`
	ProjectKey types.String             `tfsdk:"project_key"`
	Rules      []restrictionRuleModel   `tfsdk:"restriction"`
}

// restrictionRuleModel represents a single branch restriction rule.
type restrictionRuleModel struct {
	Type        types.String `tfsdk:"type"`
	MatcherType types.String `tfsdk:"matcher_type"`
	MatcherID   types.String `tfsdk:"matcher_id"`
	Users       types.Set    `tfsdk:"users"`
	Groups      types.Set    `tfsdk:"groups"`
}

// Metadata returns the resource type name.
func (r *branchPermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_branch_permissions"
}

// Schema defines the schema for the resource.
func (r *branchPermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages branch restriction rules for a Bitbucket Data Center project. " +
			"Defines which users and groups are exempted from each restriction type on matching branches.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the resource (equal to project_key).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_key": schema.StringAttribute{
				Description: "The key of the project to manage branch restrictions for. Forces replacement if changed.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"restriction": schema.SetNestedBlock{
				Description: "A branch restriction rule. Each block defines one restriction of a given type " +
					"applied to branches matching the specified matcher.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "The type of restriction. One of: read-only, no-deletes, fast-forward-only, pull-request-only.",
							Required:    true,
							Validators: []validator.String{
								&branchRestrictionTypeValidator{},
							},
						},
						"matcher_type": schema.StringAttribute{
							Description: "The type of branch matcher. One of: BRANCH, PATTERN, MODEL_CATEGORY, MODEL_BRANCH, ANY_REF.",
							Required:    true,
							Validators: []validator.String{
								&branchMatcherTypeValidator{},
							},
						},
						"matcher_id": schema.StringAttribute{
							Description: "The matcher identifier. For BRANCH: branch name (e.g. main). " +
								"For PATTERN: glob pattern (e.g. release/*). " +
								"For MODEL_CATEGORY: BUGFIX, FEATURE, HOTFIX, or RELEASE. " +
								"For MODEL_BRANCH: production or development. " +
								"For ANY_REF: use ANY_REF_MATCHER_ID.",
							Required: true,
						},
						"users": schema.SetAttribute{
							Description: "Usernames (slugs) exempt from this restriction. " +
								"Omit or set to [] to apply the restriction to everyone.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"groups": schema.SetAttribute{
							Description: "Group names exempt from this restriction. " +
								"Omit or set to [] to apply the restriction to everyone.",
							Optional:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *branchPermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *branchPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan branchPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating branch permissions", map[string]interface{}{
		"project_key": plan.ProjectKey.ValueString(),
		"rules":       len(plan.Rules),
	})

	if err := r.applyRestrictions(ctx, plan.ProjectKey.ValueString(), plan.Rules); err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Create Branch Permissions", err)...)
		return
	}

	// Read back the full state from the API.
	rules, err := r.listRestrictions(ctx, plan.ProjectKey.ValueString())
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Read Branch Permissions After Create", err)...)
		return
	}

	plan.ID = types.StringValue(plan.ProjectKey.ValueString())
	plan.Rules = rules

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *branchPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state branchPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading branch permissions", map[string]interface{}{
		"project_key": state.ProjectKey.ValueString(),
	})

	rules, err := r.listRestrictions(ctx, state.ProjectKey.ValueString())
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Read Branch Permissions", err)...)
		return
	}

	state.ID = types.StringValue(state.ProjectKey.ValueString())
	state.Rules = rules

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update updates the resource using full reconciliation.
func (r *branchPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan branchPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating branch permissions", map[string]interface{}{
		"project_key": plan.ProjectKey.ValueString(),
	})

	// Read current restrictions from API to get server IDs for deletion.
	current, err := r.listRestrictionsWithIDs(ctx, plan.ProjectKey.ValueString())
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Read Current Branch Permissions", err)...)
		return
	}

	// Build desired map keyed by semantic key.
	desired := make(map[string]restrictionRuleModel, len(plan.Rules))
	for _, rule := range plan.Rules {
		desired[branchPermissionKey(rule)] = rule
	}

	// Build current map keyed by semantic key.
	currentMap := make(map[string]restrictionWithID)
	for _, item := range current {
		currentMap[item.key] = item
	}

	authCtx := r.client.NewAuthContext(ctx)

	// Delete restrictions no longer desired or whose users/groups changed.
	for key, item := range currentMap {
		desiredRule, stillWanted := desired[key]
		if !stillWanted || !ruleUsersGroupsEqual(item.rule, desiredRule) {
			idStr := fmt.Sprintf("%d", item.id)
			httpResp, delErr := r.client.GetAPIClient().ProjectAPI.
				DeleteRestriction(authCtx, plan.ProjectKey.ValueString(), idStr).Execute()
			if delErr != nil && (httpResp == nil || httpResp.StatusCode != http.StatusNotFound) {
				resp.Diagnostics.Append(client.HandleError("Failed to Delete Branch Restriction", client.ParseErrorResponse(httpResp))...)
				return
			}
		}
	}

	// Create restrictions that are new or were deleted due to changes.
	for key, rule := range desired {
		item, exists := currentMap[key]
		if !exists || !ruleUsersGroupsEqual(item.rule, rule) {
			if err := r.createOneRestriction(ctx, plan.ProjectKey.ValueString(), rule); err != nil {
				resp.Diagnostics.Append(client.HandleError("Failed to Create Branch Restriction", err)...)
				return
			}
		}
	}

	// Read back full state.
	rules, err := r.listRestrictions(ctx, plan.ProjectKey.ValueString())
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Read Branch Permissions After Update", err)...)
		return
	}

	plan.ID = types.StringValue(plan.ProjectKey.ValueString())
	plan.Rules = rules

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete removes all branch restrictions managed by this resource.
func (r *branchPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state branchPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting branch permissions", map[string]interface{}{
		"project_key": state.ProjectKey.ValueString(),
	})

	current, err := r.listRestrictionsWithIDs(ctx, state.ProjectKey.ValueString())
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to List Branch Restrictions for Delete", err)...)
		return
	}

	authCtx := r.client.NewAuthContext(ctx)
	for _, item := range current {
		idStr := fmt.Sprintf("%d", item.id)
		httpResp, delErr := r.client.GetAPIClient().ProjectAPI.
			DeleteRestriction(authCtx, state.ProjectKey.ValueString(), idStr).Execute()
		if delErr != nil && (httpResp == nil || httpResp.StatusCode != http.StatusNotFound) {
			resp.Diagnostics.Append(client.HandleError("Failed to Delete Branch Restriction", client.ParseErrorResponse(httpResp))...)
			return
		}
	}
}

// ImportState imports the resource into Terraform state.
// Import ID format: {project_key}
func (r *branchPermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectKey := strings.TrimSpace(req.ID)
	if projectKey == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be the project key (e.g. MYPROJ)",
		)
		return
	}

	rules, err := r.listRestrictions(ctx, projectKey)
	if err != nil {
		resp.Diagnostics.Append(client.HandleError("Failed to Import Branch Permissions", err)...)
		return
	}

	state := branchPermissionsResourceModel{
		ID:         types.StringValue(projectKey),
		ProjectKey: types.StringValue(projectKey),
		Rules:      rules,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ── Helpers ───────────────────────────────────────────────────────────────

type restrictionWithID struct {
	id   int32
	key  string
	rule restrictionRuleModel
}

// listRestrictions returns all restrictions for a project as Terraform models.
func (r *branchPermissionsResource) listRestrictions(ctx context.Context, projectKey string) ([]restrictionRuleModel, error) {
	items, err := r.listRestrictionsWithIDs(ctx, projectKey)
	if err != nil {
		return nil, err
	}
	rules := make([]restrictionRuleModel, 0, len(items))
	for _, item := range items {
		rules = append(rules, item.rule)
	}
	return rules, nil
}

// listRestrictionsWithIDs returns all restrictions including their server-assigned IDs.
func (r *branchPermissionsResource) listRestrictionsWithIDs(ctx context.Context, projectKey string) ([]restrictionWithID, error) {
	authCtx := r.client.NewAuthContext(ctx)
	var result []restrictionWithID
	var start float32

	for {
		page, httpResp, err := r.client.GetAPIClient().ProjectAPI.
			GetRestrictions(authCtx, projectKey).
			Start(start).Limit(100).Execute()
		if err != nil {
			return nil, client.ParseErrorResponse(httpResp)
		}

		for _, restr := range page.GetValues() {
			rule := mapRestrictionToModel(restr)
			key := branchPermissionKey(rule)
			result = append(result, restrictionWithID{
				id:   restr.GetId(),
				key:  key,
				rule: rule,
			})
		}

		if page.GetIsLastPage() {
			break
		}
		start = float32(page.GetNextPageStart())
	}

	return result, nil
}

// applyRestrictions creates all the desired restrictions (used on Create).
func (r *branchPermissionsResource) applyRestrictions(ctx context.Context, projectKey string, rules []restrictionRuleModel) error {
	for _, rule := range rules {
		if err := r.createOneRestriction(ctx, projectKey, rule); err != nil {
			return err
		}
	}
	return nil
}

// createOneRestriction creates a single branch restriction via the API.
func (r *branchPermissionsResource) createOneRestriction(ctx context.Context, projectKey string, rule restrictionRuleModel) error {
	authCtx := r.client.NewAuthContext(ctx)

	userSlugs := setToStrings(rule.Users)
	groupNames := setToStrings(rule.Groups)

	matcherTypeID := rule.MatcherType.ValueString()
	matcherID := rule.MatcherID.ValueString()

	matcherType := bitbucket.NewUpdatePullRequestCondition1RequestSourceMatcherType(
		matcherTypeID,
		matcherTypeName(matcherTypeID),
	)
	matcher := bitbucket.NewUpdatePullRequestCondition1RequestSourceMatcherWithDefaults()
	matcher.SetId(matcherID)
	matcher.SetDisplayId(matcherDisplayID(matcherTypeID, matcherID))
	matcher.SetType(*matcherType)

	reqBody := bitbucket.NewRestRestrictionRequest([]int32{}, groupNames, userSlugs)
	reqBody.SetType(rule.Type.ValueString())
	reqBody.SetMatcher(*matcher)

	_, httpResp, err := r.client.GetAPIClient().ProjectAPI.
		CreateRestrictions(authCtx, projectKey).
		RestRestrictionRequest([]bitbucket.RestRestrictionRequest{*reqBody}).
		Execute()
	if err != nil {
		// The Bitbucket API returns a JSON array when the request is an array,
		// but the generated client expects a single object. If the HTTP status
		// is 2xx the restriction was created successfully — ignore the parse error.
		if httpResp != nil && httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
			return nil
		}
		return client.ParseErrorResponse(httpResp)
	}
	return nil
}

// mapRestrictionToModel converts an API RestRefRestriction to a Terraform model.
func mapRestrictionToModel(restr bitbucket.RestRefRestriction) restrictionRuleModel {
	matcherType := ""
	matcherID := ""
	if m, ok := restr.GetMatcherOk(); ok {
		matcherID = m.GetId()
		if mt, ok2 := m.GetTypeOk(); ok2 {
			matcherType = mt.GetId()
		}
	}

	// Extract user slugs. Normalize empty → null so Optional (non-Computed) attributes
	// in SetNestedBlock produce consistent hashes between plan and state.
	var usersSet types.Set
	if apiUsers := restr.GetUsers(); len(apiUsers) > 0 {
		userSlugs := make([]attr.Value, len(apiUsers))
		for i, u := range apiUsers {
			userSlugs[i] = types.StringValue(u.GetSlug())
		}
		usersSet, _ = types.SetValue(types.StringType, userSlugs)
	} else {
		usersSet = types.SetNull(types.StringType)
	}

	// Extract group names. Normalize empty → null for the same reason.
	var groupsSet types.Set
	if apiGroups := restr.GetGroups(); len(apiGroups) > 0 {
		groupNames := make([]attr.Value, len(apiGroups))
		for i, g := range apiGroups {
			groupNames[i] = types.StringValue(g)
		}
		groupsSet, _ = types.SetValue(types.StringType, groupNames)
	} else {
		groupsSet = types.SetNull(types.StringType)
	}

	return restrictionRuleModel{
		Type:        types.StringValue(restr.GetType()),
		MatcherType: types.StringValue(matcherType),
		MatcherID:   types.StringValue(matcherID),
		Users:       usersSet,
		Groups:      groupsSet,
	}
}

// branchPermissionKey returns a stable key for a restriction rule to use in reconciliation.
func branchPermissionKey(r restrictionRuleModel) string {
	return r.Type.ValueString() + ":" + r.MatcherType.ValueString() + ":" + r.MatcherID.ValueString()
}

// ruleUsersGroupsEqual returns true if two rules have equivalent users and groups.
// null and an empty set are treated as equivalent since both mean "no exemptions".
func ruleUsersGroupsEqual(a, b restrictionRuleModel) bool {
	return normalizeStringSet(a.Users).Equal(normalizeStringSet(b.Users)) &&
		normalizeStringSet(a.Groups).Equal(normalizeStringSet(b.Groups))
}

// normalizeStringSet converts null/empty sets to a canonical empty set for comparison.
func normalizeStringSet(s types.Set) types.Set {
	if s.IsNull() || s.IsUnknown() || len(s.Elements()) == 0 {
		empty, _ := types.SetValue(types.StringType, []attr.Value{})
		return empty
	}
	return s
}

// setToStrings extracts string values from a types.Set.
func setToStrings(s types.Set) []string {
	if s.IsNull() || s.IsUnknown() {
		return []string{}
	}
	result := make([]string, 0, len(s.Elements()))
	for _, elem := range s.Elements() {
		if sv, ok := elem.(types.String); ok {
			result = append(result, sv.ValueString())
		}
	}
	return result
}

// matcherTypeName maps a Bitbucket matcher type ID to its human-readable display name.
func matcherTypeName(id string) string {
	switch id {
	case "BRANCH":
		return "Branch"
	case "PATTERN":
		return "Pattern"
	case "MODEL_CATEGORY":
		return "Model category"
	case "MODEL_BRANCH":
		return "Model branch"
	case "ANY_REF":
		return "Any ref"
	default:
		return id
	}
}

// matcherDisplayID returns the display ID for a given matcher type and ID.
func matcherDisplayID(matcherType, matcherID string) string {
	if matcherType == "ANY_REF" {
		return "ANY (wildcard)"
	}
	return matcherID
}

// ── Validators ────────────────────────────────────────────────────────────

var validRestrictionTypes = map[string]bool{
	"read-only":          true,
	"no-deletes":         true,
	"fast-forward-only":  true,
	"pull-request-only":  true,
}

type branchRestrictionTypeValidator struct{}

func (v *branchRestrictionTypeValidator) Description(_ context.Context) string {
	return "Branch restriction type must be one of: read-only, no-deletes, fast-forward-only, pull-request-only"
}

func (v *branchRestrictionTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *branchRestrictionTypeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !validRestrictionTypes[req.ConfigValue.ValueString()] {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Branch Restriction Type",
			fmt.Sprintf("%q is not a valid restriction type. Must be one of: read-only, no-deletes, fast-forward-only, pull-request-only.", req.ConfigValue.ValueString()),
		)
	}
}

var validMatcherTypes = map[string]bool{
	"BRANCH":         true,
	"PATTERN":        true,
	"MODEL_CATEGORY": true,
	"MODEL_BRANCH":   true,
	"ANY_REF":        true,
}

type branchMatcherTypeValidator struct{}

func (v *branchMatcherTypeValidator) Description(_ context.Context) string {
	return "Branch matcher type must be one of: BRANCH, PATTERN, MODEL_CATEGORY, MODEL_BRANCH, ANY_REF"
}

func (v *branchMatcherTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *branchMatcherTypeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !validMatcherTypes[req.ConfigValue.ValueString()] {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Branch Matcher Type",
			fmt.Sprintf("%q is not a valid matcher type. Must be one of: BRANCH, PATTERN, MODEL_CATEGORY, MODEL_BRANCH, ANY_REF.", req.ConfigValue.ValueString()),
		)
	}
}
