# Project Permissions Resource

## Overview

The `bitbucketdc_project_permissions` resource manages user and group permissions for a Bitbucket Data Center project. This resource handles the complete set of permissions for a project, adding and removing users/groups as needed to match the desired state.

## Resource Schema

### Required Attributes

- `project_key` (String) - The key of the project to manage permissions for.

### Optional Nested Blocks

- `user` (Block Set) - Set of user permissions. Each block contains:
  - `name` (String, Required) - The username
  - `permission` (String, Required) - Permission level: `PROJECT_READ`, `PROJECT_WRITE`, or `PROJECT_ADMIN`

- `group` (Block Set) - Set of group permissions. Each block contains:
  - `name` (String, Required) - The group name
  - `permission` (String, Required) - Permission level: `PROJECT_READ`, `PROJECT_WRITE`, or `PROJECT_ADMIN`

### Computed Attributes

- `id` (String) - The unique identifier (same as project_key).

## Terraform Configuration Example

```hcl
# Basic permissions
resource "bitbucketdc_project_permissions" "example" {
  project_key = bitbucketdc_project.example.key

  user {
    name       = "john.doe"
    permission = "PROJECT_ADMIN"
  }

  user {
    name       = "jane.smith"
    permission = "PROJECT_WRITE"
  }

  group {
    name       = "developers"
    permission = "PROJECT_WRITE"
  }

  group {
    name       = "viewers"
    permission = "PROJECT_READ"
  }
}

# Using data sources
resource "bitbucketdc_project_permissions" "with_datasources" {
  project_key = bitbucketdc_project.example.key

  user {
    name       = data.bitbucketdc_user.admin.username
    permission = "PROJECT_ADMIN"
  }

  group {
    name       = data.bitbucketdc_group.team.name
    permission = "PROJECT_WRITE"
  }
}

# Admin-only project
resource "bitbucketdc_project_permissions" "locked_down" {
  project_key = bitbucketdc_project.private.key

  user {
    name       = "admin"
    permission = "PROJECT_ADMIN"
  }
}
```

## Import

Existing project permissions can be imported using the project key:

```bash
terraform import bitbucketdc_project_permissions.example MYPROJ
```

**Note**: Import will read all current permissions and store them in state. You'll need to update your configuration to match the imported state or run `terraform apply` to reconcile differences.

## API Mapping

### Create/Update Operations
The resource performs multiple API calls to reconcile state:

1. **List Current Permissions**:
   - User Permissions: `GET /rest/api/1.0/projects/{projectKey}/permissions/users`
   - Group Permissions: `GET /rest/api/1.0/projects/{projectKey}/permissions/groups`

2. **Grant Permissions**:
   - User: `PUT /rest/api/1.0/projects/{projectKey}/permissions/users?name={username}&permission={level}`
   - Group: `PUT /rest/api/1.0/projects/{projectKey}/permissions/groups?name={groupname}&permission={level}`

3. **Revoke Permissions**:
   - User: `DELETE /rest/api/1.0/projects/{projectKey}/permissions/users?name={username}`
   - Group: `DELETE /rest/api/1.0/projects/{projectKey}/permissions/groups?name={groupname}`

### Read Operation
- Lists all users and groups with permissions
- Populates state with current permission levels

### Delete Operation
- Revokes all permissions defined in the resource
- **Warning**: May leave project inaccessible if deleting all admin permissions

## Permission Levels

| Level | Description |
|-------|-------------|
| `PROJECT_READ` | Can view project and clone repositories |
| `PROJECT_WRITE` | Can push to repositories and create branches |
| `PROJECT_ADMIN` | Full control: manage settings, permissions, hooks |

## Validation Rules

1. **Username/Group Name**:
   - Must exist in Bitbucket
   - Validated via API during apply

2. **Permission Level**:
   - Must be one of: `PROJECT_READ`, `PROJECT_WRITE`, `PROJECT_ADMIN`
   - Case-sensitive

3. **Uniqueness**:
   - Each user can appear only once in the configuration
   - Each group can appear only once in the configuration
   - Terraform will error on duplicate users/groups in config

## State Behavior

### Reconciliation Logic

The resource performs a three-way reconciliation:

1. **Add Missing**: Grant permissions for users/groups in config but not in Bitbucket
2. **Update Changed**: Update permission level for users/groups with different levels
3. **Remove Extra**: Revoke permissions for users/groups in Bitbucket but not in config

### Example Reconciliation

**Current State (Bitbucket)**:
- john.doe: PROJECT_ADMIN
- jane.smith: PROJECT_READ

**Desired State (Terraform)**:
- john.doe: PROJECT_ADMIN (no change)
- jane.smith: PROJECT_WRITE (update)
- bob.jones: PROJECT_READ (add)

**Actions Performed**:
1. Update jane.smith to PROJECT_WRITE
2. Grant bob.jones PROJECT_READ
3. (john.doe unchanged)

### State Refresh
- Read operation lists all current permissions
- Updates state to match Bitbucket
- Drift detection shows manual changes made outside Terraform

## Error Handling

### Common Errors

1. **User Not Found**:
   - API Error: 404 Not Found
   - Terraform Error: "Failed to Grant Permission: User 'john.doe' does not exist in Bitbucket. Verify the username is correct."

2. **Group Not Found**:
   - API Error: 404 Not Found
   - Terraform Error: "Failed to Grant Permission: Group 'developers' does not exist in Bitbucket. Create the group or correct the name."

3. **Insufficient Permissions**:
   - API Error: 403 Forbidden
   - Terraform Error: "Failed to Manage Permissions: Your credentials do not have PROJECT_ADMIN permission for project 'MYPROJ'."

4. **Last Admin Removal**:
   - API Error: 400 Bad Request (may vary)
   - Terraform Error: "Failed to Revoke Permission: Cannot remove last admin from project 'MYPROJ'. At least one PROJECT_ADMIN user or group must remain."

5. **Project Not Found**:
   - API Error: 404 Not Found
   - Terraform Error: "Failed to Manage Permissions: Project 'MYPROJ' does not exist. Ensure the project is created first."

## Implementation Details

### Resource Structure

```go
type projectPermissionsResourceModel struct {
    ID         types.String `tfsdk:"id"`
    ProjectKey types.String `tfsdk:"project_key"`
    Users      []userPermissionModel  `tfsdk:"user"`
    Groups     []groupPermissionModel `tfsdk:"group"`
}

type userPermissionModel struct {
    Name       types.String `tfsdk:"name"`
    Permission types.String `tfsdk:"permission"`
}

type groupPermissionModel struct {
    Name       types.String `tfsdk:"name"`
    Permission types.String `tfsdk:"permission"`
}
```

### Schema Definition

```go
func (r *projectPermissionsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "Manages user and group permissions for a Bitbucket Data Center project.",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed:    true,
                Description: "The unique identifier (same as project_key).",
            },
            "project_key": schema.StringAttribute{
                Required:    true,
                Description: "The key of the project to manage permissions for.",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
        },
        Blocks: map[string]schema.Block{
            "user": schema.SetNestedBlock{
                Description: "User permissions for the project.",
                NestedObject: schema.NestedBlockObject{
                    Attributes: map[string]schema.Attribute{
                        "name": schema.StringAttribute{
                            Required:    true,
                            Description: "The username.",
                        },
                        "permission": schema.StringAttribute{
                            Required:    true,
                            Description: "Permission level: PROJECT_READ, PROJECT_WRITE, or PROJECT_ADMIN.",
                            Validators: []validator.String{
                                stringvalidator.OneOf("PROJECT_READ", "PROJECT_WRITE", "PROJECT_ADMIN"),
                            },
                        },
                    },
                },
            },
            "group": schema.SetNestedBlock{
                Description: "Group permissions for the project.",
                NestedObject: schema.NestedBlockObject{
                    Attributes: map[string]schema.Attribute{
                        "name": schema.StringAttribute{
                            Required:    true,
                            Description: "The group name.",
                        },
                        "permission": schema.StringAttribute{
                            Required:    true,
                            Description: "Permission level: PROJECT_READ, PROJECT_WRITE, or PROJECT_ADMIN.",
                            Validators: []validator.String{
                                stringvalidator.OneOf("PROJECT_READ", "PROJECT_WRITE", "PROJECT_ADMIN"),
                            },
                        },
                    },
                },
            },
        },
    }
}
```

### Update Implementation (Reconciliation)

```go
func (r *projectPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan, state projectPermissionsResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    projectKey := plan.ProjectKey.ValueString()

    // Get current permissions from API
    currentUsers, err := r.client.ListProjectUserPermissions(ctx, projectKey)
    if err != nil {
        resp.Diagnostics.AddError("Failed to Read Permissions", err.Error())
        return
    }

    currentGroups, err := r.client.ListProjectGroupPermissions(ctx, projectKey)
    if err != nil {
        resp.Diagnostics.AddError("Failed to Read Permissions", err.Error())
        return
    }

    // Reconcile users
    for _, desiredUser := range plan.Users {
        username := desiredUser.Name.ValueString()
        permission := desiredUser.Permission.ValueString()

        currentPerm, exists := findUserPermission(currentUsers, username)
        if !exists {
            // Grant new permission
            err := r.client.GrantUserPermission(ctx, projectKey, username, permission)
            if err != nil {
                resp.Diagnostics.AddError("Failed to Grant Permission", err.Error())
                return
            }
        } else if currentPerm != permission {
            // Update existing permission
            err := r.client.GrantUserPermission(ctx, projectKey, username, permission)
            if err != nil {
                resp.Diagnostics.AddError("Failed to Update Permission", err.Error())
                return
            }
        }
    }

    // Revoke users not in desired state
    for _, currentUser := range currentUsers {
        if !isUserInPlan(plan.Users, currentUser.Name) {
            err := r.client.RevokeUserPermission(ctx, projectKey, currentUser.Name)
            if err != nil {
                resp.Diagnostics.AddError("Failed to Revoke Permission", err.Error())
                return
            }
        }
    }

    // Similar logic for groups...

    // Read back final state
    resp.Diagnostics.Append(r.Read(ctx, &readReq, &readResp)...)
}
```

## Testing Strategy

### Acceptance Tests

```hcl
# Test: Basic permissions
resource "bitbucketdc_project" "test" {
  key  = "TEST"
  name = "Test Project"
}

resource "bitbucketdc_project_permissions" "test" {
  project_key = bitbucketdc_project.test.key

  user {
    name       = "testuser"
    permission = "PROJECT_ADMIN"
  }

  group {
    name       = "testgroup"
    permission = "PROJECT_WRITE"
  }
}

# Test: Update permissions
# Update the config above to change permission levels and verify reconciliation

# Test: Remove permissions
# Remove user/group from config and verify they are revoked
```

### Test Cases
1. Create permissions for new project
2. Add new user to existing permissions
3. Add new group to existing permissions
4. Update user permission level
5. Update group permission level
6. Remove user from permissions
7. Remove group from permissions
8. Import existing permissions
9. Handle non-existent user error
10. Handle non-existent group error
11. Handle insufficient permissions error
12. Verify reconciliation removes extra permissions
13. Verify reconciliation adds missing permissions
14. Verify reconciliation updates changed permissions

## Dependencies

### Required Before
- Project must exist (bitbucketdc_project)
- Users must exist in Bitbucket
- Groups must exist in Bitbucket

### Optional Data Sources
- `bitbucketdc_user` - Query user details
- `bitbucketdc_group` - Query group details

## Rollback Strategy

- If creation/update fails, state remains unchanged
- User can retry operation
- Manual cleanup may be required if partial application occurred
- Consider using `terraform refresh` to sync state with current permissions

## Security Considerations

- Requires `PROJECT_ADMIN` permission to manage permissions
- Removing all admins may lock you out of the project
- Use caution when revoking permissions for critical users
- Consider using groups for permission management (easier to maintain)
- Audit logs in Bitbucket track permission changes

## Best Practices

1. **Use Groups**: Prefer group permissions over individual user permissions for easier management
2. **Maintain Admin Access**: Always ensure at least one admin user/group has access
3. **Separate Resources**: Consider using separate permission resources for different permission tiers
4. **Data Sources**: Use data sources to reference users and groups for validation
5. **Import First**: When adopting existing projects, import permissions before making changes

## Performance Considerations

- Multiple API calls required for reconciliation
- Performance degrades with large number of users/groups
- Consider batching permission changes where possible
- Refresh operations list all permissions (can be slow for large projects)

## Alternative Approaches

### Fine-Grained Resources
Instead of managing all permissions in one resource, consider separate resources for granular control:

```hcl
# Alternative: Per-user permission resources
resource "bitbucketdc_project_user_permission" "admin" {
  project_key = bitbucketdc_project.example.key
  username    = "john.doe"
  permission  = "PROJECT_ADMIN"
}

resource "bitbucketdc_project_user_permission" "writer" {
  project_key = bitbucketdc_project.example.key
  username    = "jane.smith"
  permission  = "PROJECT_WRITE"
}
```

**Trade-offs**:
- Pro: More granular state management, smaller blast radius
- Con: More resources to manage, more complex dependencies

## Future Enhancements

- Support for default permissions (template)
- Support for permission inheritance
- Bulk import of permissions from CSV/JSON
- Permission diff preview before apply
- Lifecycle policy to prevent removing last admin
- Support for conditional permissions based on external factors
