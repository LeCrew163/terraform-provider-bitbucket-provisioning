## Why

Teams provisioning Bitbucket projects via Terraform need to reference and manage the groups and users that receive project permissions — but the provider currently offers no way to create groups, manage users, or control group membership. Without these, permission resources (`bitbucketdc_project_permissions`, `bitbucketdc_repository_permissions`) can only reference groups and users that were created manually outside of Terraform.

## What Changes

- Add `bitbucketdc_group` resource — create and manage Bitbucket user groups
- Add `bitbucketdc_user` resource — create and manage local Bitbucket users
- Add `bitbucketdc_user_group` resource — manage membership of a user in a group

## Capabilities

### New Capabilities
- `group-resource`: Create, update, and delete Bitbucket user groups
- `user-resource`: Create, update, and delete local Bitbucket users
- `user-group-resource`: Manage user membership within a group (add/remove)

### Modified Capabilities

## Impact

- New files: `resource_group.go`, `resource_user.go`, `resource_user_group.go` and their acceptance tests
- `provider.go` updated to register three new resources
- Enables fully declarative team onboarding: create group → create users → assign to group → grant project permissions — all in one Terraform configuration
