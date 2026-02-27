## ADDED Requirements

### Requirement: Create and delete a Bitbucket user group
The system SHALL provide a `bitbucketdc_group` resource that creates and deletes a Bitbucket DC user group via the PermissionManagement API.

#### Scenario: Create a group
- **WHEN** user declares `resource "bitbucketdc_group" "devs" { name = "developers" }`
- **THEN** system calls `CreateGroup(name)` and the group is created in Bitbucket

#### Scenario: Delete a group
- **WHEN** user runs `terraform destroy` on a `bitbucketdc_group` resource
- **THEN** system calls `DeleteGroup(name)` and the group is removed from Bitbucket

#### Scenario: Import a group by name
- **WHEN** user runs `terraform import bitbucketdc_group.devs developers`
- **THEN** system sets resource ID and name to `developers` and state is populated

#### Scenario: Rename forces replacement
- **WHEN** user changes `name` in an existing `bitbucketdc_group` resource
- **THEN** system destroys and recreates the resource (RequiresReplace) because no rename API exists

### Requirement: Group resource ID is the group name
The system SHALL use the group name as the resource ID.

#### Scenario: Group ID equals name
- **WHEN** `bitbucketdc_group.devs` is created with `name = "developers"`
- **THEN** `bitbucketdc_group.devs.id` equals `"developers"`
