# Default Reviewers Specification

## ADDED Requirements

### Requirement: Configure default reviewer rules
The system SHALL allow configuration of default reviewer rules through YAML to automatically add reviewers to pull requests.

#### Scenario: Add default reviewers for all pull requests
- **WHEN** user configures default reviewers without branch restrictions
- **THEN** system creates a rule that adds those reviewers to all new pull requests

#### Scenario: Add default reviewers for specific branches
- **WHEN** user configures default reviewers with branch pattern restrictions
- **THEN** system creates a rule that adds reviewers only to pull requests targeting matching branches

#### Scenario: Multiple reviewer rules for different branches
- **WHEN** user configures multiple default reviewer rules with different branch patterns
- **THEN** system creates all rules with appropriate branch matching

### Requirement: Support user and group reviewers
The system SHALL allow specification of both individual users and groups as default reviewers.

#### Scenario: Add individual users as default reviewers
- **WHEN** user specifies usernames in default reviewer configuration
- **THEN** system adds those users as required reviewers to matching pull requests

#### Scenario: Add groups as default reviewers
- **WHEN** user specifies group names in default reviewer configuration
- **THEN** system adds all members of those groups as reviewers to matching pull requests

#### Scenario: Mix of users and groups as reviewers
- **WHEN** user specifies both usernames and group names in reviewer configuration
- **THEN** system adds both specified users and group members as reviewers

### Requirement: Configure required reviewer count
The system SHALL allow specification of the minimum number of approvals required from default reviewers.

#### Scenario: Require specific number of approvals
- **WHEN** user specifies required approval count in reviewer rule
- **THEN** system enforces that number of approvals before allowing merge

#### Scenario: Require all reviewers to approve
- **WHEN** user sets required approval count to match number of reviewers
- **THEN** system requires approval from all default reviewers

#### Scenario: Default approval count
- **WHEN** user does not specify required approval count
- **THEN** system uses default requirement of at least one approval

### Requirement: Support source branch matching
The system SHALL allow configuration of reviewer rules based on source branch patterns.

#### Scenario: Reviewers for feature branches
- **WHEN** user configures reviewers for pull requests from branches matching "feature/*"
- **THEN** system adds those reviewers to pull requests originating from feature branches

#### Scenario: Reviewers for hotfix branches
- **WHEN** user configures reviewers for pull requests from branches matching "hotfix/*"
- **THEN** system adds those reviewers to pull requests originating from hotfix branches

### Requirement: Support target branch matching
The system SHALL allow configuration of reviewer rules based on target branch patterns.

#### Scenario: Reviewers for pull requests to main
- **WHEN** user configures reviewers for pull requests targeting main branch
- **THEN** system adds those reviewers to all pull requests targeting main

#### Scenario: Reviewers for pull requests to release branches
- **WHEN** user configures reviewers for pull requests targeting branches matching "release/*"
- **THEN** system adds those reviewers to pull requests targeting any release branch

### Requirement: Validate reviewer existence
The system SHALL validate that specified reviewers exist in Bitbucket before creating rules.

#### Scenario: Valid reviewers specified
- **WHEN** user specifies usernames and groups that exist in Bitbucket
- **THEN** system passes validation and creates reviewer rules

#### Scenario: Non-existent user specified
- **WHEN** user specifies a username that does not exist in Bitbucket
- **THEN** system fails validation and reports the non-existent username

#### Scenario: Non-existent group specified
- **WHEN** user specifies a group name that does not exist in Bitbucket
- **THEN** system fails validation and reports the non-existent group name

### Requirement: Support idempotent reviewer rule management
The system SHALL support idempotent reviewer rule operations where applying the same configuration produces the same result.

#### Scenario: Reapply unchanged reviewer rules
- **WHEN** user applies the same default reviewer configuration
- **THEN** system detects rules are already correct and makes no changes

#### Scenario: Update existing reviewer rules
- **WHEN** user modifies reviewer rules in YAML configuration
- **THEN** system updates the rules to match new configuration

#### Scenario: Remove reviewer rules
- **WHEN** user removes reviewer rules from YAML configuration
- **THEN** system deletes those rules from the project

### Requirement: Validate reviewer rule configuration
The system SHALL validate default reviewer rule syntax and values before applying changes.

#### Scenario: Valid reviewer rule configuration
- **WHEN** user provides well-formed reviewer rule configuration
- **THEN** system passes validation and applies the rules

#### Scenario: Invalid branch pattern
- **WHEN** user specifies malformed branch pattern in reviewer rule
- **THEN** system fails validation with error explaining pattern syntax

#### Scenario: Invalid approval count
- **WHEN** user specifies negative or zero approval count
- **THEN** system fails validation requiring positive approval count

### Requirement: Report reviewer rule changes
The system SHALL report all default reviewer rule changes made during configuration application.

#### Scenario: New reviewer rules created
- **WHEN** system creates new default reviewer rules
- **THEN** system reports each rule with branch patterns, reviewers, and approval count

#### Scenario: Reviewer rules updated
- **WHEN** system modifies existing reviewer rules
- **THEN** system reports each rule with old and new configuration

#### Scenario: Reviewer rules deleted
- **WHEN** system removes reviewer rules
- **THEN** system reports each rule deleted with its configuration

### Requirement: Handle reviewer rule precedence
The system SHALL handle cases where multiple reviewer rules match the same pull request.

#### Scenario: Multiple rules match same pull request
- **WHEN** multiple reviewer rules match a pull request's source and target branches
- **THEN** system applies all matching rules adding all specified reviewers

#### Scenario: Overlapping reviewers in multiple rules
- **WHEN** multiple matching rules specify the same reviewer
- **THEN** system adds the reviewer only once to the pull request

### Requirement: Support reviewer exclusions
The system SHALL allow specification of conditions that exclude certain users from being added as reviewers.

#### Scenario: Exclude pull request author
- **WHEN** user configures reviewer rule with author exclusion
- **THEN** system does not add the pull request author as a reviewer even if they match the rule

#### Scenario: Exclude specific users from rule
- **WHEN** user specifies exclusion list in reviewer rule
- **THEN** system does not add excluded users as reviewers even if they are in specified groups

### Requirement: Configure reviewer notification settings
The system SHALL allow configuration of how reviewers are notified when added to pull requests.

#### Scenario: Enable reviewer notifications
- **WHEN** user enables notifications in reviewer rule configuration
- **THEN** system sends notification emails to reviewers when added to pull requests

#### Scenario: Disable reviewer notifications
- **WHEN** user disables notifications in reviewer rule configuration
- **THEN** system adds reviewers to pull requests without sending notification emails

### Requirement: Support conditional reviewer requirements
The system SHALL allow specification of conditions under which reviewers are required vs optional.

#### Scenario: Required reviewers must approve
- **WHEN** user marks reviewers as required in rule configuration
- **THEN** system blocks merge until required number of approvals received

#### Scenario: Optional reviewers suggested but not required
- **WHEN** user marks reviewers as optional in rule configuration
- **THEN** system adds reviewers but allows merge without their approval
