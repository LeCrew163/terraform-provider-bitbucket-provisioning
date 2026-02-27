# Project Resource

## Overview

The `bitbucketdc_project` resource manages Bitbucket Data Center projects. Projects are top-level containers for repositories and provide a namespace for organizing related repositories, managing permissions, and applying shared configurations.

## Resource Schema

### Required Attributes

- `key` (String) - The project key, a unique identifier for the project. Must be uppercase alphanumeric with underscores. Cannot be changed after creation (forces replacement).
- `name` (String) - The human-readable name of the project.

### Optional Attributes

- `description` (String) - A description of the project's purpose. Defaults to empty string.
- `visibility` (String) - Project visibility, either "public" or "private". Defaults to "private".

### Computed Attributes

- `id` (String) - The unique identifier assigned by Bitbucket. Typically the same as the key.

## Terraform Configuration Example

```hcl
# Basic project
resource "bitbucketdc_project" "example" {
  key         = "MYPROJ"
  name        = "My Project"
  description = "Example project for demonstration"
  visibility  = "private"
}

# Public project
resource "bitbucketdc_project" "open_source" {
  key        = "OSS"
  name       = "Open Source Projects"
  description = "Public repositories for open source contributions"
  visibility = "public"
}

# Minimal configuration
resource "bitbucketdc_project" "minimal" {
  key  = "MIN"
  name = "Minimal Project"
}
```

## Import

Existing projects can be imported using the project key:

```bash
terraform import bitbucketdc_project.example MYPROJ
```

## API Mapping

### Create Operation
- **API Endpoint**: `POST /rest/api/1.0/projects`
- **Request Body**:
  ```json
  {
    "key": "MYPROJ",
    "name": "My Project",
    "description": "Example project",
    "public": false
  }
  ```

### Read Operation
- **API Endpoint**: `GET /rest/api/1.0/projects/{projectKey}`
- **Response**: Project object with key, name, description, and visibility

### Update Operation
- **API Endpoint**: `PUT /rest/api/1.0/projects/{projectKey}`
- **Request Body**: Same as create, but key cannot be changed
- **Notes**: Only name, description, and visibility can be updated

### Delete Operation
- **API Endpoint**: `DELETE /rest/api/1.0/projects/{projectKey}`
- **Notes**: Project must be empty (no repositories) before deletion

## Validation Rules

1. **Key Format**:
   - Must match regex: `^[A-Z][A-Z0-9_]*$`
   - Minimum length: 2 characters
   - Maximum length: 128 characters
   - Only uppercase letters, numbers, and underscores
   - Must start with a letter

2. **Name**:
   - Minimum length: 1 character
   - Maximum length: 255 characters

3. **Visibility**:
   - Must be either "public" or "private"
   - Case-sensitive

## State Behavior

### Force Replacement
- Changing `key` forces resource replacement (cannot be updated in-place)

### Computed Values
- `id` is computed from the API response after creation

### State Refresh
- Read operation refreshes state on every `terraform plan` and `terraform refresh`
- If project is deleted outside Terraform, resource is removed from state

## Error Handling

### Common Errors

1. **Duplicate Key**:
   - API Error: 409 Conflict
   - Message: "A project with key 'MYPROJ' already exists"
   - Terraform Error: "Failed to Create Project: A project with this key already exists. Choose a different key or import the existing project."

2. **Permission Denied**:
   - API Error: 403 Forbidden
   - Message: "You do not have permission to create projects"
   - Terraform Error: "Failed to Create Project: Insufficient permissions. Ensure your credentials have PROJECT_CREATE permission."

3. **Invalid Key Format**:
   - Caught by validator before API call
   - Terraform Error: "Invalid value for key: must be uppercase alphanumeric with underscores, starting with a letter"

4. **Project Not Empty (Delete)**:
   - API Error: 400 Bad Request
   - Message: "Project must be empty to delete"
   - Terraform Error: "Failed to Delete Project: Project contains repositories. Remove all repositories before deleting the project."

## Implementation Details

### Resource Structure

```go
type projectResourceModel struct {
    ID          types.String `tfsdk:"id"`
    Key         types.String `tfsdk:"key"`
    Name        types.String `tfsdk:"name"`
    Description types.String `tfsdk:"description"`
    Visibility  types.String `tfsdk:"visibility"`
}
```

### Schema Definition

```go
func (r *projectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "Manages a Bitbucket Data Center project.",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed:    true,
                Description: "The unique identifier of the project.",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "key": schema.StringAttribute{
                Required:    true,
                Description: "The project key (uppercase alphanumeric with underscores).",
                Validators: []validator.String{
                    validators.ProjectKeyValidator(),
                },
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "name": schema.StringAttribute{
                Required:    true,
                Description: "The name of the project.",
            },
            "description": schema.StringAttribute{
                Optional:    true,
                Computed:    true,
                Default:     stringdefault.StaticString(""),
                Description: "The description of the project.",
            },
            "visibility": schema.StringAttribute{
                Optional:    true,
                Computed:    true,
                Default:     stringdefault.StaticString("private"),
                Description: "The visibility of the project (public or private).",
                Validators: []validator.String{
                    stringvalidator.OneOf("public", "private"),
                },
            },
        },
    }
}
```

### Create Implementation

```go
func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data projectResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Call API to create project
    project, err := r.client.CreateProject(ctx, &bitbucket.CreateProjectRequest{
        Key:         data.Key.ValueString(),
        Name:        data.Name.ValueString(),
        Description: data.Description.ValueString(),
        Public:      data.Visibility.ValueString() == "public",
    })

    if err != nil {
        resp.Diagnostics.AddError(
            "Failed to Create Project",
            fmt.Sprintf("Could not create project '%s': %s", data.Key.ValueString(), err),
        )
        return
    }

    // Set computed values
    data.ID = types.StringValue(project.Key)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

## Testing Strategy

### Unit Tests
- Test schema validation (key format, visibility values)
- Test state model transformations
- Test import ID parsing

### Acceptance Tests

```hcl
# Test: Basic project creation
resource "bitbucketdc_project" "test" {
  key  = "TEST"
  name = "Test Project"
}

# Test: Full configuration
resource "bitbucketdc_project" "full" {
  key         = "FULL"
  name        = "Full Project"
  description = "Complete configuration"
  visibility  = "public"
}

# Test: Import
# terraform import bitbucketdc_project.import IMPORT
```

### Test Cases
1. Create project with minimal configuration
2. Create project with full configuration
3. Update project name
4. Update project description
5. Update project visibility (private to public)
6. Import existing project
7. Handle duplicate key error
8. Handle invalid key format
9. Handle permission denied error
10. Force replacement when key changes

## Dependencies

### Required Before
- Provider configuration (authentication)

### Required After
- Repositories within the project
- Project permissions
- Project configurations (hooks, branch permissions, etc.)

## Rollback Strategy

- If creation fails, no state is saved
- If update fails, state remains unchanged
- If delete fails, state remains (manual intervention required)
- User can retry failed operations

## Security Considerations

- Project creation requires `PROJECT_CREATE` system permission
- Only users with `PROJECT_ADMIN` permission can update/delete projects
- Credentials are marked sensitive in provider configuration
- No sensitive data stored in project resource state

## Performance Considerations

- Project operations are typically fast (< 1 second)
- No pagination required for single project operations
- Read operations are called on every plan/refresh
- Consider using data sources for read-only project references

## Future Enhancements

- Support for project avatar configuration
- Support for project type (Normal vs Personal)
- Support for custom project settings
- Lifecycle policies for preventing accidental deletion
