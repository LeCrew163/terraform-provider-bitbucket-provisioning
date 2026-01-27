# Repository Resource

## Overview

The `bitbucketdc_repository` resource manages Git repositories within Bitbucket Data Center projects. Repositories store code and provide version control capabilities.

## Resource Schema

### Required Attributes

- `project_key` (String) - The key of the project containing this repository. Forces replacement if changed.
- `slug` (String) - The repository slug (URL-friendly identifier). Must be lowercase alphanumeric with hyphens. Forces replacement if changed.
- `name` (String) - The human-readable name of the repository.

### Optional Attributes

- `description` (String) - A description of the repository's purpose. Defaults to empty string.
- `forkable` (Boolean) - Whether the repository can be forked. Defaults to true.
- `public` (Boolean) - Whether the repository is publicly accessible. Defaults to false. Note: Project visibility affects this.
- `default_branch` (String) - The default branch name. Computed from Bitbucket if not specified.

### Computed Attributes

- `id` (String) - The unique identifier (format: `{project_key}/{slug}`).
- `clone_url_http` (String) - The HTTP(S) clone URL.
- `clone_url_ssh` (String) - The SSH clone URL.
- `state` (String) - Repository state (e.g., "AVAILABLE").

## Terraform Configuration Example

```hcl
# Basic repository
resource "bitbucketdc_repository" "example" {
  project_key = bitbucketdc_project.example.key
  slug        = "my-repo"
  name        = "My Repository"
}

# Full configuration
resource "bitbucketdc_repository" "full" {
  project_key = bitbucketdc_project.example.key
  slug        = "full-repo"
  name        = "Full Repository"
  description = "Repository with all options configured"
  forkable    = true
  public      = false
  default_branch = "main"
}

# Reference existing project
resource "bitbucketdc_repository" "with_datasource" {
  project_key = data.bitbucketdc_project.existing.key
  slug        = "new-repo"
  name        = "New Repository"
}
```

## Import

Existing repositories can be imported using the format `{project_key}/{slug}`:

```bash
terraform import bitbucketdc_repository.example MYPROJ/my-repo
```

## API Mapping

### Create Operation
- **API Endpoint**: `POST /rest/api/1.0/projects/{projectKey}/repos`
- **Request Body**:
  ```json
  {
    "name": "My Repository",
    "scmId": "git",
    "forkable": true,
    "public": false
  }
  ```

### Read Operation
- **API Endpoint**: `GET /rest/api/1.0/projects/{projectKey}/repos/{repositorySlug}`
- **Response**: Repository object with slug, name, description, links, etc.

### Update Operation
- **API Endpoint**: `PUT /rest/api/1.0/projects/{projectKey}/repos/{repositorySlug}`
- **Request Body**: Updated repository properties
- **Notes**: slug and project cannot be changed

### Delete Operation
- **API Endpoint**: `DELETE /rest/api/1.0/projects/{projectKey}/repos/{repositorySlug}`
- **Notes**: Repository must be empty or force delete parameter required

## Validation Rules

1. **Slug Format**:
   - Must match regex: `^[a-z0-9][a-z0-9-]*[a-z0-9]$`
   - Minimum length: 1 character
   - Maximum length: 128 characters
   - Only lowercase letters, numbers, and hyphens
   - Cannot start or end with hyphen

2. **Name**:
   - Minimum length: 1 character
   - Maximum length: 255 characters

3. **Project Key**:
   - Must reference existing project
   - Validated via API on create

## State Behavior

### Force Replacement
- Changing `project_key` forces resource replacement
- Changing `slug` forces resource replacement

### Computed Values
- `id` is computed as `{project_key}/{slug}`
- `clone_url_http` and `clone_url_ssh` are computed from API response
- `state` is computed from API response
- `default_branch` is computed if not specified

### State Refresh
- Read operation refreshes state on every plan/refresh
- If repository is deleted outside Terraform, resource is removed from state

## Error Handling

### Common Errors

1. **Duplicate Slug**:
   - API Error: 409 Conflict
   - Terraform Error: "Failed to Create Repository: A repository with slug 'my-repo' already exists in project 'MYPROJ'. Choose a different slug or import the existing repository."

2. **Project Not Found**:
   - API Error: 404 Not Found
   - Terraform Error: "Failed to Create Repository: Project 'MYPROJ' does not exist. Ensure the project is created first."

3. **Permission Denied**:
   - API Error: 403 Forbidden
   - Terraform Error: "Failed to Create Repository: Insufficient permissions. Ensure your credentials have REPO_CREATE permission for project 'MYPROJ'."

4. **Invalid Slug Format**:
   - Caught by validator before API call
   - Terraform Error: "Invalid value for slug: must be lowercase alphanumeric with hyphens, cannot start or end with hyphen"

## Implementation Details

### Resource Structure

```go
type repositoryResourceModel struct {
    ID            types.String `tfsdk:"id"`
    ProjectKey    types.String `tfsdk:"project_key"`
    Slug          types.String `tfsdk:"slug"`
    Name          types.String `tfsdk:"name"`
    Description   types.String `tfsdk:"description"`
    Forkable      types.Bool   `tfsdk:"forkable"`
    Public        types.Bool   `tfsdk:"public"`
    DefaultBranch types.String `tfsdk:"default_branch"`
    CloneURLHTTP  types.String `tfsdk:"clone_url_http"`
    CloneURLSSH   types.String `tfsdk:"clone_url_ssh"`
    State         types.String `tfsdk:"state"`
}
```

### Schema Definition

```go
func (r *repositoryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "Manages a Bitbucket Data Center repository.",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed:    true,
                Description: "The unique identifier of the repository (format: project_key/slug).",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "project_key": schema.StringAttribute{
                Required:    true,
                Description: "The key of the project containing this repository.",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "slug": schema.StringAttribute{
                Required:    true,
                Description: "The repository slug (lowercase alphanumeric with hyphens).",
                Validators: []validator.String{
                    validators.RepositorySlugValidator(),
                },
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "name": schema.StringAttribute{
                Required:    true,
                Description: "The name of the repository.",
            },
            "description": schema.StringAttribute{
                Optional:    true,
                Computed:    true,
                Default:     stringdefault.StaticString(""),
                Description: "The description of the repository.",
            },
            "forkable": schema.BoolAttribute{
                Optional:    true,
                Computed:    true,
                Default:     booldefault.StaticBool(true),
                Description: "Whether the repository can be forked.",
            },
            "public": schema.BoolAttribute{
                Optional:    true,
                Computed:    true,
                Default:     booldefault.StaticBool(false),
                Description: "Whether the repository is publicly accessible.",
            },
            "default_branch": schema.StringAttribute{
                Optional:    true,
                Computed:    true,
                Description: "The default branch name.",
            },
            "clone_url_http": schema.StringAttribute{
                Computed:    true,
                Description: "The HTTP(S) clone URL.",
            },
            "clone_url_ssh": schema.StringAttribute{
                Computed:    true,
                Description: "The SSH clone URL.",
            },
            "state": schema.StringAttribute{
                Computed:    true,
                Description: "The repository state.",
            },
        },
    }
}
```

## Testing Strategy

### Acceptance Tests

```hcl
# Test: Basic repository creation
resource "bitbucketdc_project" "test" {
  key  = "TEST"
  name = "Test Project"
}

resource "bitbucketdc_repository" "test" {
  project_key = bitbucketdc_project.test.key
  slug        = "test-repo"
  name        = "Test Repository"
}

# Test: Full configuration
resource "bitbucketdc_repository" "full" {
  project_key    = bitbucketdc_project.test.key
  slug           = "full-repo"
  name           = "Full Repository"
  description    = "Complete configuration"
  forkable       = true
  public         = false
  default_branch = "main"
}
```

### Test Cases
1. Create repository with minimal configuration
2. Create repository with full configuration
3. Update repository name
4. Update repository description
5. Update forkable setting
6. Import existing repository
7. Handle duplicate slug error
8. Handle invalid slug format
9. Handle missing project error
10. Force replacement when project_key or slug changes
11. Verify clone URLs are populated
12. Verify state is computed

## Dependencies

### Required Before
- Project must exist (bitbucketdc_project)

### Required After
- Repository permissions
- Repository hooks
- Branch configurations

## Rollback Strategy

- If creation fails, no state is saved
- If update fails, state remains unchanged
- If delete fails, state remains (manual intervention required)
- User can retry failed operations

## Security Considerations

- Repository creation requires `REPO_CREATE` permission in the project
- Only users with `REPO_ADMIN` permission can update/delete repositories
- Public repositories are accessible to all users (including unauthenticated)
- SSH keys for clone access managed separately via access keys resource

## Performance Considerations

- Repository operations are typically fast (< 2 seconds)
- Initial clone operations (outside Terraform) may be slow for large repos
- Read operations are called on every plan/refresh
- Use data sources for read-only repository references

## Future Enhancements

- Support for forking existing repositories
- Support for repository mirroring
- Support for repository archiving
- Support for pull request settings
- Integration with default branch protection rules
