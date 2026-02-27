# Access Keys Specification

## ADDED Requirements

### Requirement: Add SSH access keys to project
The system SHALL allow addition of SSH public keys to a project for read-only access through YAML configuration.

#### Scenario: Add new SSH key with label
- **WHEN** user specifies an SSH public key and label in YAML configuration
- **THEN** system adds the key to the project with the specified label

#### Scenario: Add multiple SSH keys
- **WHEN** user specifies multiple SSH public keys in YAML configuration
- **THEN** system adds all specified keys to the project

#### Scenario: SSH key already exists
- **WHEN** user applies configuration with an SSH key that already exists in the project
- **THEN** system detects the duplicate and skips addition or updates the label if changed

### Requirement: Remove SSH access keys from project
The system SHALL allow removal of SSH access keys through YAML configuration updates.

#### Scenario: Remove specific SSH key
- **WHEN** user removes an SSH key from YAML configuration
- **THEN** system deletes that key from the project

#### Scenario: Remove all SSH keys
- **WHEN** user removes all SSH keys from YAML configuration
- **THEN** system deletes all access keys from the project

### Requirement: Support SSH key permissions
The system SHALL configure appropriate permissions for SSH access keys based on their intended use.

#### Scenario: Read-only access key
- **WHEN** user adds an SSH key with read-only permission specification
- **THEN** system creates an access key allowing clone and fetch operations only

#### Scenario: Read-write access key
- **WHEN** user adds an SSH key with read-write permission specification
- **THEN** system creates an access key allowing clone, fetch, and push operations

### Requirement: Validate SSH key format
The system SHALL validate SSH public key format before attempting to add keys to a project.

#### Scenario: Valid SSH public key format
- **WHEN** user provides a properly formatted SSH public key
- **THEN** system passes validation and adds the key

#### Scenario: Invalid SSH key format
- **WHEN** user provides malformed SSH public key
- **THEN** system fails validation with error explaining required SSH key format

#### Scenario: Unsupported key type
- **WHEN** user provides SSH key with unsupported algorithm
- **THEN** system fails validation and lists supported SSH key types

### Requirement: Label access keys for identification
The system SHALL require meaningful labels for each SSH access key to identify their purpose.

#### Scenario: Unique label for each key
- **WHEN** user provides unique labels for each SSH key
- **THEN** system creates keys with the specified labels for easy identification

#### Scenario: Missing label
- **WHEN** user does not provide a label for an SSH key
- **THEN** system fails validation requiring a label for each key

#### Scenario: Duplicate label
- **WHEN** user provides the same label for multiple SSH keys
- **THEN** system fails validation requiring unique labels for each key

### Requirement: Support SSH key sources
The system SHALL allow SSH public keys to be specified inline or loaded from files.

#### Scenario: Inline SSH key in YAML
- **WHEN** user provides SSH public key directly in YAML configuration
- **THEN** system uses the inline key value

#### Scenario: SSH key from file reference
- **WHEN** user provides a file path to an SSH public key in YAML configuration
- **THEN** system reads the key from the specified file and adds it

#### Scenario: SSH key file does not exist
- **WHEN** user references a non-existent SSH key file
- **THEN** system fails validation with error indicating file not found

### Requirement: Support idempotent key management
The system SHALL support idempotent access key operations where applying the same configuration produces the same result.

#### Scenario: Reapply unchanged key configuration
- **WHEN** user applies the same access key configuration
- **THEN** system detects keys are already correct and makes no changes

#### Scenario: Update key label
- **WHEN** user changes a key's label in YAML configuration
- **THEN** system updates the label without regenerating the key

### Requirement: Report access key changes
The system SHALL report all access key changes made during configuration application.

#### Scenario: New keys added
- **WHEN** system adds new access keys
- **THEN** system reports each key added with its label and fingerprint

#### Scenario: Keys removed
- **WHEN** system removes access keys
- **THEN** system reports each key removed with its label

#### Scenario: Key labels updated
- **WHEN** system updates key labels
- **THEN** system reports each key with old and new labels

### Requirement: Display key fingerprints
The system SHALL display SSH key fingerprints for verification and auditing purposes.

#### Scenario: Show fingerprint after adding key
- **WHEN** system successfully adds an SSH key
- **THEN** system displays the key's fingerprint in the output

#### Scenario: List existing keys with fingerprints
- **WHEN** user requests to show current project state
- **THEN** system displays all access keys with their labels and fingerprints

### Requirement: Handle key conflicts
The system SHALL handle cases where the same SSH key exists in multiple locations or with different labels.

#### Scenario: Same key added with different label
- **WHEN** user attempts to add an SSH key that already exists with a different label
- **THEN** system detects the duplicate key and reports the conflict

#### Scenario: Key exists at user level and project level
- **WHEN** an SSH key exists both as a user key and project key
- **THEN** system allows both and documents that user keys and project keys are independent
