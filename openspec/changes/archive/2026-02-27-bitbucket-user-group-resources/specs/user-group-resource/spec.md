## ADDED Requirements

### Requirement: Manage user membership in a Bitbucket group
The system SHALL provide a `bitbucketdc_user_group` resource that adds and removes a user from a Bitbucket DC group.

#### Scenario: Add a user to a group
- **WHEN** user declares `resource "bitbucketdc_user_group" "membership" { username = "alice"; group = "developers" }`
- **THEN** system calls `AddUserToGroup` with `itemName = alice` and `context = developers`

#### Scenario: Remove a user from a group
- **WHEN** user runs `terraform destroy` on a `bitbucketdc_user_group` resource
- **THEN** system calls `RemoveUserFromGroup` with query params `user = alice` and `context = developers`

#### Scenario: Import membership by username/group
- **WHEN** user runs `terraform import bitbucketdc_user_group.m alice/developers`
- **THEN** system sets resource ID to `alice/developers` and verifies membership exists

#### Scenario: Out-of-band removal handled gracefully
- **WHEN** user is removed from the group outside of Terraform
- **THEN** next `terraform plan` shows the resource as missing and proposes to recreate it (no crash)

### Requirement: Resource ID is username/group composite
The system SHALL use `username/group` as the resource ID.

#### Scenario: ID format
- **WHEN** `bitbucketdc_user_group.m` is created with `username = "alice"` and `group = "developers"`
- **THEN** `bitbucketdc_user_group.m.id` equals `"alice/developers"`

### Requirement: All attributes require replacement on change
The system SHALL mark `username` and `group` as RequiresReplace because there is no update operation.

#### Scenario: Changing username forces replacement
- **WHEN** user changes `username` on an existing membership resource
- **THEN** Terraform destroys the old membership and creates a new one

#### Scenario: Changing group forces replacement
- **WHEN** user changes `group` on an existing membership resource
- **THEN** Terraform destroys the old membership and creates a new one
