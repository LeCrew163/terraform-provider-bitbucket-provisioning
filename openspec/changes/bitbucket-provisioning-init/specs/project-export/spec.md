# Project Export Specification

## ADDED Requirements

### Requirement: Export project configuration to YAML
The system SHALL export existing Bitbucket project configuration to YAML format for migration and version control.

#### Scenario: Export project to file
- **WHEN** user executes export command with project key and output file path
- **THEN** system creates YAML file containing all project configuration

#### Scenario: Export project to stdout
- **WHEN** user executes export command with project key without output file
- **THEN** system outputs YAML configuration to stdout

#### Scenario: Project does not exist
- **WHEN** user attempts to export a non-existent project key
- **THEN** system fails with clear error message indicating project not found

### Requirement: Export all project capabilities
The system SHALL export configuration for all 7 capabilities in a single unified YAML file.

#### Scenario: Export includes project settings
- **WHEN** user exports a project
- **THEN** output includes project key, name, description, visibility, and avatar

#### Scenario: Export includes permissions
- **WHEN** user exports a project with configured permissions
- **THEN** output includes all user and group permissions

#### Scenario: Export includes branch permissions
- **WHEN** user exports a project with branch permission rules
- **THEN** output includes all branch protection rules and restrictions

#### Scenario: Export includes access keys
- **WHEN** user exports a project with SSH access keys
- **THEN** output includes all access keys with labels and permissions

#### Scenario: Export includes workflow configuration
- **WHEN** user exports a project with branching model configured
- **THEN** output includes workflow type and branch configuration

#### Scenario: Export includes hooks
- **WHEN** user exports a project with configured hooks
- **THEN** output includes all enabled hooks with their settings

#### Scenario: Export includes default reviewers
- **WHEN** user exports a project with default reviewer rules
- **THEN** output includes all reviewer rules with patterns and requirements

### Requirement: Generate valid YAML compatible with apply command
The system SHALL generate YAML that can be directly used with the apply command without modification.

#### Scenario: Exported YAML passes validation
- **WHEN** user exports a project and validates the generated YAML
- **THEN** validation succeeds without errors

#### Scenario: Exported YAML can be applied
- **WHEN** user exports a project and applies the exported YAML to a new project
- **THEN** new project is created with identical configuration

#### Scenario: Export-modify-apply workflow
- **WHEN** user exports a project, modifies the YAML, and applies it back
- **THEN** system updates only the modified fields

### Requirement: Include metadata in export
The system SHALL include metadata about the export operation for auditing and troubleshooting.

#### Scenario: Export includes timestamp
- **WHEN** user exports a project
- **THEN** YAML includes export timestamp in metadata

#### Scenario: Export includes tool version
- **WHEN** user exports a project
- **THEN** YAML includes bitbucket-provisioning tool version in metadata

#### Scenario: Export includes Bitbucket version
- **WHEN** user exports a project
- **THEN** YAML includes Bitbucket Data Center version in metadata

#### Scenario: Export includes source project key
- **WHEN** user exports a project
- **THEN** YAML includes original project key in metadata

### Requirement: Format YAML for readability
The system SHALL format exported YAML for human readability with proper structure and comments.

#### Scenario: YAML uses proper indentation
- **WHEN** user exports a project
- **THEN** output uses consistent 2-space indentation

#### Scenario: YAML includes helpful comments
- **WHEN** user exports a project
- **THEN** output includes comments indicating optional fields and descriptions

#### Scenario: YAML groups related configuration
- **WHEN** user exports a project
- **THEN** output groups capabilities in logical sections

#### Scenario: Empty capabilities are omitted
- **WHEN** user exports a project with some capabilities not configured
- **THEN** output omits empty sections for unconfigured capabilities

### Requirement: Handle sensitive data appropriately
The system SHALL handle sensitive data in exports with appropriate warnings and masking.

#### Scenario: Warn about sensitive data
- **WHEN** user exports a project containing sensitive data (tokens, secrets)
- **THEN** system warns that exported file may contain sensitive information

#### Scenario: Mask webhook secrets
- **WHEN** user exports a project with webhooks containing secrets
- **THEN** output indicates secrets are present but does not include plaintext values

#### Scenario: Include SSH key fingerprints only
- **WHEN** user exports a project with SSH access keys
- **THEN** output includes key labels and permissions but not the private key content

### Requirement: Support verbose output mode
The system SHALL provide verbose output showing what is being exported.

#### Scenario: Verbose mode shows capability discovery
- **WHEN** user exports with verbose flag enabled
- **THEN** system reports each capability being read from Bitbucket

#### Scenario: Verbose mode shows API calls
- **WHEN** user exports with verbose flag enabled
- **THEN** system reports each API endpoint being queried

#### Scenario: Verbose mode shows skipped capabilities
- **WHEN** user exports with verbose flag and some capabilities are unconfigured
- **THEN** system reports which capabilities are being skipped

### Requirement: Validate user permissions before export
The system SHALL validate that the user has sufficient permissions to read project configuration.

#### Scenario: User has read access
- **WHEN** user with project read permission exports a project
- **THEN** system successfully exports all readable configuration

#### Scenario: User lacks read access
- **WHEN** user without project read permission attempts export
- **THEN** system fails with clear error about insufficient permissions

#### Scenario: User has partial permissions
- **WHEN** user with limited permissions exports a project
- **THEN** system exports accessible capabilities and notes which are inaccessible

### Requirement: Support JSON output format
The system SHALL support exporting configuration in JSON format as alternative to YAML.

#### Scenario: Export to JSON format
- **WHEN** user specifies JSON format flag
- **THEN** system outputs configuration in valid JSON format

#### Scenario: JSON export for programmatic use
- **WHEN** user exports to JSON format
- **THEN** output is valid JSON parseable by standard JSON libraries

### Requirement: Handle export errors gracefully
The system SHALL provide clear error messages when export fails.

#### Scenario: API error during export
- **WHEN** Bitbucket API returns error during export
- **THEN** system reports which capability failed and the error details

#### Scenario: Partial export on non-critical failure
- **WHEN** export of one capability fails but others succeed
- **THEN** system exports available data and reports which capability failed

#### Scenario: File write permission error
- **WHEN** user specifies output file in directory without write permissions
- **THEN** system fails with clear error about file permissions

### Requirement: Support project discovery
The system SHALL provide ability to list available projects for export.

#### Scenario: List all projects
- **WHEN** user requests project list
- **THEN** system displays all accessible project keys and names

#### Scenario: List projects with filter
- **WHEN** user requests project list with search pattern
- **THEN** system displays matching projects only

#### Scenario: List shows export readiness
- **WHEN** user requests project list
- **THEN** system indicates which projects are ready to export

### Requirement: Provide export diff preview
The system SHALL show what would be exported before writing to file.

#### Scenario: Preview export without creating file
- **WHEN** user uses preview flag with export command
- **THEN** system displays what would be exported without creating output file

#### Scenario: Confirm before overwriting file
- **WHEN** user exports to existing file without force flag
- **THEN** system prompts for confirmation before overwriting

### Requirement: Support batch export operations
The system SHALL allow exporting multiple projects in single operation.

#### Scenario: Export multiple projects by pattern
- **WHEN** user specifies multiple project keys or pattern
- **THEN** system exports each project to separate files

#### Scenario: Export all projects
- **WHEN** user requests export of all accessible projects
- **THEN** system creates YAML file for each project

#### Scenario: Batch export with progress indicator
- **WHEN** user exports multiple projects
- **THEN** system shows progress for each project being exported
