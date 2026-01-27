# Project Permissions Specification

## ADDED Requirements

### Requirement: Manage user permissions at project level
The system SHALL allow configuration of individual user permissions at the project level through YAML definition.

#### Scenario: Grant user project admin permission
- **WHEN** user specifies a username with PROJECT_ADMIN permission in YAML
- **THEN** system grants that user admin rights to the project

#### Scenario: Grant user project write permission
- **WHEN** user specifies a username with PROJECT_WRITE permission in YAML
- **THEN** system grants that user write access to the project

#### Scenario: Grant user project read permission
- **WHEN** user specifies a username with PROJECT_READ permission in YAML
- **THEN** system grants that user read-only access to the project

#### Scenario: Remove user permission
- **WHEN** user removes a username from permissions in updated YAML configuration
- **THEN** system revokes that user's explicit project permissions

### Requirement: Manage group permissions at project level
The system SHALL allow configuration of group permissions at the project level through YAML definition.

#### Scenario: Grant group project admin permission
- **WHEN** user specifies a group name with PROJECT_ADMIN permission in YAML
- **THEN** system grants all members of that group admin rights to the project

#### Scenario: Grant group project write permission
- **WHEN** user specifies a group name with PROJECT_WRITE permission in YAML
- **THEN** system grants all members of that group write access to the project

#### Scenario: Grant group project read permission
- **WHEN** user specifies a group name with PROJECT_READ permission in YAML
- **THEN** system grants all members of that group read-only access to the project

#### Scenario: Remove group permission
- **WHEN** user removes a group from permissions in updated YAML configuration
- **THEN** system revokes that group's explicit project permissions

### Requirement: Validate user existence before granting permissions
The system SHALL validate that users exist in Bitbucket before attempting to grant permissions.

#### Scenario: Valid user exists
- **WHEN** user specifies a username that exists in Bitbucket
- **THEN** system proceeds to grant the specified permissions

#### Scenario: User does not exist
- **WHEN** user specifies a username that does not exist in Bitbucket
- **THEN** system fails validation and reports the non-existent username

### Requirement: Validate group existence before granting permissions
The system SHALL validate that groups exist in Bitbucket before attempting to grant permissions.

#### Scenario: Valid group exists
- **WHEN** user specifies a group name that exists in Bitbucket
- **THEN** system proceeds to grant the specified permissions

#### Scenario: Group does not exist
- **WHEN** user specifies a group name that does not exist in Bitbucket
- **THEN** system fails validation and reports the non-existent group name

### Requirement: Support idempotent permission management
The system SHALL support idempotent permission operations where applying the same configuration produces the same result.

#### Scenario: Reapply unchanged permissions
- **WHEN** user applies the same permission configuration to a project
- **THEN** system detects permissions are already correct and makes no changes

#### Scenario: Update existing permissions
- **WHEN** user changes a user's permission level in YAML configuration
- **THEN** system updates that user's permission to the new level

#### Scenario: Add new permissions without affecting existing
- **WHEN** user adds new users or groups to permission configuration
- **THEN** system grants new permissions while preserving existing ones

### Requirement: Support multiple permission levels
The system SHALL support all Bitbucket project permission levels as defined by the API.

#### Scenario: Assign PROJECT_ADMIN level
- **WHEN** user specifies PROJECT_ADMIN permission level
- **THEN** system grants full administrative access including project settings and permissions management

#### Scenario: Assign PROJECT_WRITE level
- **WHEN** user specifies PROJECT_WRITE permission level
- **THEN** system grants write access to repositories without administrative privileges

#### Scenario: Assign PROJECT_READ level
- **WHEN** user specifies PROJECT_READ permission level
- **THEN** system grants read-only access to project repositories

### Requirement: Handle permission conflicts
The system SHALL handle cases where users have conflicting permissions from multiple sources.

#### Scenario: User has both direct and group permissions
- **WHEN** user has permissions both directly and through group membership
- **THEN** system applies the highest permission level among all sources

#### Scenario: Multiple group memberships with different permissions
- **WHEN** user belongs to multiple groups with different permission levels
- **THEN** system results in the user having the highest permission level from all groups

### Requirement: Report permission changes
The system SHALL report all permission changes made during application of configuration.

#### Scenario: New permissions granted
- **WHEN** system grants new permissions
- **THEN** system reports each user and group added with their permission level

#### Scenario: Permissions updated
- **WHEN** system changes existing permission levels
- **THEN** system reports each user and group updated with old and new permission levels

#### Scenario: Permissions revoked
- **WHEN** system removes permissions
- **THEN** system reports each user and group whose permissions were revoked

### Requirement: Validate permission syntax in YAML
The system SHALL validate permission configuration syntax and values before attempting to apply changes.

#### Scenario: Valid permission configuration
- **WHEN** user provides well-formed permission configuration with valid permission levels
- **THEN** system passes validation and proceeds to apply permissions

#### Scenario: Invalid permission level
- **WHEN** user specifies an invalid permission level not recognized by Bitbucket
- **THEN** system fails validation and lists valid permission levels

#### Scenario: Malformed permission structure
- **WHEN** user provides incorrectly structured permission configuration in YAML
- **THEN** system fails validation with clear error indicating the structural issue
