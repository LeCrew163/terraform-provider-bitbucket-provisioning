## ADDED Requirements

### Requirement: Create and delete a local Bitbucket user
The system SHALL provide a `bitbucketdc_user` resource that creates and deletes a local Bitbucket DC user via the PermissionManagement API.

#### Scenario: Create a user
- **WHEN** user declares a `bitbucketdc_user` resource with `name`, `display_name`, `email_address`, and `password`
- **THEN** system calls `CreateUser` and the user account is created in Bitbucket

#### Scenario: Delete a user
- **WHEN** user runs `terraform destroy` on a `bitbucketdc_user` resource
- **THEN** system calls `DeleteUser(name)` and the account is removed from Bitbucket

#### Scenario: Import a user by username slug
- **WHEN** user runs `terraform import bitbucketdc_user.alice alice`
- **THEN** system sets the resource ID to `alice` and state is populated (password left empty)

### Requirement: Username can be updated in-place via RenameUser
The system SHALL support renaming a user without forcing resource replacement.

#### Scenario: Rename a user
- **WHEN** user changes `name` on an existing `bitbucketdc_user` resource
- **THEN** system calls `RenameUser(oldName, newName)` and the resource ID is updated to the new name

### Requirement: Password is write-only and sensitive
The system SHALL mark the `password` attribute as sensitive and never read it back from the API.

#### Scenario: Password not exposed in plan output
- **WHEN** user reads `bitbucketdc_user.alice.password`
- **THEN** the value is marked sensitive and does not appear in plain text in plan or state output

#### Scenario: Password drift not detected
- **WHEN** a user's password is changed outside of Terraform
- **THEN** Terraform plan shows no diff (password is write-only, not reconciled)

### Requirement: Optional user creation flags
The system SHALL support `add_to_default_group` and `notify` boolean attributes on create.

#### Scenario: Add to default group on creation
- **WHEN** user sets `add_to_default_group = true`
- **THEN** Bitbucket adds the user to its default group upon creation

#### Scenario: Send email notification on creation
- **WHEN** user sets `notify = true`
- **THEN** Bitbucket sends a welcome email to the user's email address
