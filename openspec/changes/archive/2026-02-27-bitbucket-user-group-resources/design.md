## Context

The generated API client (`PermissionManagementAPIService`) already covers all required operations: `CreateGroup`, `DeleteGroup`, `CreateUser`, `DeleteUser`, `RenameUser`, `AddUserToGroup`, and `RemoveUserFromGroup`. No raw HTTP workarounds are needed. The existing provider patterns (resource CRUD + import) apply directly.

## Goals / Non-Goals

**Goals:**
- `bitbucketdc_group` — create/delete a Bitbucket user group; import by group name
- `bitbucketdc_user` — create/delete a local Bitbucket user; import by username (slug)
- `bitbucketdc_user_group` — add/remove a user from a group; import by `username/group`

**Non-Goals:**
- Global permission management (requires Bitbucket admin role — out of scope)
- LDAP/external directory users (read-only in Bitbucket DC, cannot be created via API)
- User password rotation after creation (Bitbucket API provides no update-password endpoint)

## Decisions

### Group resource: name is the ID, no update
`CreateGroup` accepts only a `name`. There is no rename operation for groups. The resource ID is the group name; any name change forces replacement (`RequiresReplace`).

### User resource: password is write-only
`CreateUser` accepts `name`, `displayName`, `emailAddress`, `password`, `addToDefaultGroup`, `notify`. The password is never returned by the API, so it must be marked `Sensitive: true` and tracked only in state as written. Username can be renamed via `RenameUser` without forcing replacement — this is the only in-place update.

### user_group resource: join table pattern
ID is `username/group`. Create = `AddUserToGroup`, Delete = `RemoveUserFromGroup`. There is no update — all attributes are `RequiresReplace`. Read is via `FindGroupsForUser` to verify membership still exists (handle out-of-band removal gracefully).

### API method for AddUserToGroup
`AddUserToGroup` uses a `UserPickerContext` body with `itemName` (username) and `context` (group name) fields. `RemoveUserFromGroup` uses query params `user` (username) and `context` (group name).

## Risks / Trade-offs

- **Admin-only APIs**: `CreateGroup`, `CreateUser`, `DeleteUser`, `DeleteGroup` require Bitbucket admin (`SYS_ADMIN` or `ADMIN`). Users who don't have this role will get 403 errors at apply time. This is a Bitbucket constraint, not a provider bug — document it clearly.
- **Password drift**: Password is write-only. If someone changes a user's password outside Terraform, the provider has no way to detect or reconcile the drift. This matches how every other Terraform provider handles credentials.
- **Group rename not supported**: Bitbucket DC has no rename-group API. Changing the name destroys and recreates — this will break any project/repo permissions referencing the old group name.
