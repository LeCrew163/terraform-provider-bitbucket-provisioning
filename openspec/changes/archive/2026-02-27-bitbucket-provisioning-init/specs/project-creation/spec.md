# Project Creation Specification

## ADDED Requirements

### Requirement: Create project from YAML configuration
The system SHALL create a new Bitbucket project based on a YAML configuration file containing project key, name, description, and settings.

#### Scenario: Create new project successfully
- **WHEN** user provides a valid YAML configuration with project key, name, and description
- **THEN** system creates the project in Bitbucket with the specified attributes

#### Scenario: Project key already exists
- **WHEN** user attempts to create a project with a key that already exists in Bitbucket
- **THEN** system validates current state and updates the project if configuration differs

#### Scenario: Invalid project key format
- **WHEN** user provides a project key that does not meet Bitbucket naming requirements
- **THEN** system fails validation and returns a clear error message explaining the requirements

### Requirement: Validate YAML configuration before creation
The system SHALL validate the YAML configuration against a defined schema before attempting to create the project.

#### Scenario: Valid YAML structure
- **WHEN** user provides a YAML file with all required fields and correct structure
- **THEN** system passes validation and proceeds to creation

#### Scenario: Missing required fields
- **WHEN** user provides a YAML file missing required fields like project key or name
- **THEN** system fails validation and lists all missing required fields

#### Scenario: Invalid YAML syntax
- **WHEN** user provides a YAML file with syntax errors
- **THEN** system fails parsing and returns the line number and error description

### Requirement: Support idempotent project creation
The system SHALL support idempotent operations where running the same configuration multiple times produces the same result.

#### Scenario: Reapply unchanged configuration
- **WHEN** user applies the same YAML configuration to an existing project with matching state
- **THEN** system detects no changes and completes successfully without API calls

#### Scenario: Reapply with configuration changes
- **WHEN** user applies an updated YAML configuration to an existing project
- **THEN** system detects differences and updates only the changed attributes

### Requirement: Perform pre-flight connectivity check
The system SHALL validate connectivity to the Bitbucket instance and verify credentials before attempting project creation.

#### Scenario: Successful connectivity check
- **WHEN** user executes a project creation command
- **THEN** system validates that Bitbucket URL is reachable and credentials are valid before proceeding

#### Scenario: Bitbucket instance unreachable
- **WHEN** user executes a project creation command and Bitbucket URL is not reachable
- **THEN** system fails immediately with a clear error message about connectivity

#### Scenario: Invalid credentials
- **WHEN** user executes a project creation command with invalid authentication credentials
- **THEN** system fails immediately with a clear error message about authentication failure

### Requirement: Support dry-run mode for preview
The system SHALL provide a dry-run mode that shows what changes would be made without actually applying them.

#### Scenario: Dry-run for new project
- **WHEN** user executes project creation with --dry-run flag
- **THEN** system displays what project would be created without making API calls

#### Scenario: Dry-run for existing project update
- **WHEN** user executes project apply with --dry-run flag on existing project
- **THEN** system displays a diff showing what would change without making modifications

### Requirement: Configure project visibility
The system SHALL allow configuration of project visibility as public or private in the YAML definition.

#### Scenario: Create private project
- **WHEN** user specifies visibility as private in YAML configuration
- **THEN** system creates a private project accessible only to users with explicit permissions

#### Scenario: Create public project
- **WHEN** user specifies visibility as public in YAML configuration
- **THEN** system creates a public project visible to all authenticated users

#### Scenario: Default visibility when not specified
- **WHEN** user does not specify visibility in YAML configuration
- **THEN** system creates a private project as the secure default

### Requirement: Set project avatar
The system SHALL allow optional configuration of a project avatar through a URL or file path.

#### Scenario: Set avatar from URL
- **WHEN** user provides an avatar URL in YAML configuration
- **THEN** system downloads the image and sets it as the project avatar

#### Scenario: Set avatar from local file
- **WHEN** user provides a local file path for avatar in YAML configuration
- **THEN** system uploads the image file and sets it as the project avatar

#### Scenario: Avatar not specified
- **WHEN** user does not specify an avatar in YAML configuration
- **THEN** system creates the project with the default Bitbucket avatar

### Requirement: Provide detailed error messages
The system SHALL provide clear, actionable error messages when project creation fails.

#### Scenario: API error with details
- **WHEN** Bitbucket API returns an error during project creation
- **THEN** system displays the error message, HTTP status code, and suggests corrective actions

#### Scenario: Permission denied error
- **WHEN** authenticated user lacks permission to create projects
- **THEN** system displays a clear message indicating insufficient permissions and required role

### Requirement: Output operation results
The system SHALL output the result of project creation operations in both human-readable and machine-readable formats.

#### Scenario: Successful creation in interactive mode
- **WHEN** user creates a project in interactive CLI mode
- **THEN** system displays a success message with project key, name, and URL

#### Scenario: Successful creation in CI/CD mode
- **WHEN** user creates a project with --json flag
- **THEN** system outputs structured JSON with project details and operation status

#### Scenario: Failed creation in CI/CD mode
- **WHEN** project creation fails with --json flag
- **THEN** system outputs structured JSON with error details and non-zero exit code

### Requirement: Support modular YAML configuration with file references
The system SHALL support referencing external YAML files for capability configurations using path-based references.

#### Scenario: Reference external file with relative path
- **WHEN** user specifies a capability configuration as a relative file path (e.g., ./configs/permissions.yaml)
- **THEN** system loads and parses the referenced file relative to the main config file location

#### Scenario: Reference external file with absolute path
- **WHEN** user specifies a capability configuration as an absolute file path (e.g., /opt/shared/hooks.yaml)
- **THEN** system loads and parses the referenced file from the absolute path

#### Scenario: Mix inline and referenced configurations
- **WHEN** user provides some capabilities inline and others as file references
- **THEN** system processes both inline configs and referenced files correctly

#### Scenario: Referenced file does not exist
- **WHEN** user references a YAML file that does not exist
- **THEN** system fails validation with error indicating which file is missing

### Requirement: Validate referenced configuration files
The system SHALL validate all referenced configuration files exist and contain valid YAML before processing.

#### Scenario: All referenced files exist and valid
- **WHEN** user references multiple external configuration files that all exist and are valid
- **THEN** system successfully loads and validates all configurations

#### Scenario: Referenced file has invalid YAML syntax
- **WHEN** user references a file with invalid YAML syntax
- **THEN** system fails validation with error indicating file path and syntax issue

#### Scenario: Referenced file contains invalid configuration
- **WHEN** user references a file with valid YAML but invalid configuration schema
- **THEN** system fails validation with error indicating file path and schema violations

### Requirement: Prevent circular file references
The system SHALL detect and prevent circular references between configuration files.

#### Scenario: Direct circular reference
- **WHEN** file A references file B, and file B references file A
- **THEN** system detects the circular reference and fails with clear error message

#### Scenario: Indirect circular reference
- **WHEN** file A references file B, file B references file C, and file C references file A
- **THEN** system detects the indirect circular reference and fails with clear error message

### Requirement: Support shared configuration reuse
The system SHALL allow multiple projects to reference the same shared configuration files.

#### Scenario: Multiple projects use shared permissions config
- **WHEN** multiple project configs reference the same permissions file
- **THEN** system loads the shared file independently for each project without conflicts

#### Scenario: Shared config in common location
- **WHEN** user references configurations from a shared directory structure
- **THEN** system resolves paths correctly and loads shared configurations

### Requirement: Preserve capability separation in referenced files
The system SHALL maintain clear separation of capabilities when using referenced files.

#### Scenario: One capability per file
- **WHEN** user creates separate files for each capability (permissions.yaml, hooks.yaml, etc.)
- **THEN** system loads each file for its respective capability only

#### Scenario: Multiple capabilities in one file
- **WHEN** user references a file containing multiple capability configurations
- **THEN** system loads all capabilities from the file and applies them appropriately

### Requirement: Report file loading and parsing
The system SHALL provide clear feedback when loading and parsing referenced configuration files.

#### Scenario: Verbose mode shows file loading
- **WHEN** user runs command with verbose flag and has file references
- **THEN** system reports each file being loaded with its path

#### Scenario: Error shows which file failed
- **WHEN** parsing fails for a referenced file
- **THEN** system error message includes the file path that failed and reason
