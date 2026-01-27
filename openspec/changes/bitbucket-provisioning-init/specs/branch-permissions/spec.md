# Branch Permissions Specification

## ADDED Requirements

### Requirement: Configure branch permission rules
The system SHALL allow configuration of branch permission rules through YAML to restrict actions on specific branch patterns.

#### Scenario: Restrict writes to main branch
- **WHEN** user defines a rule preventing writes to main branch except for specific users
- **THEN** system creates a branch permission rule enforcing the restriction

#### Scenario: Require pull requests for protected branches
- **WHEN** user defines a rule requiring pull requests for branches matching a pattern
- **THEN** system creates a branch permission rule preventing direct pushes

#### Scenario: Allow only specific users to delete branches
- **WHEN** user defines a rule restricting branch deletion to specific users
- **THEN** system creates a branch permission rule allowing only those users to delete matching branches

### Requirement: Support branch pattern matching
The system SHALL support wildcard patterns and exact branch names for branch permission rules.

#### Scenario: Exact branch name match
- **WHEN** user specifies an exact branch name like "main" in a rule
- **THEN** system applies the rule only to that specific branch

#### Scenario: Wildcard pattern match
- **WHEN** user specifies a pattern like "release/*" in a rule
- **THEN** system applies the rule to all branches matching the pattern

#### Scenario: Multiple branch patterns in one rule
- **WHEN** user specifies multiple branch patterns in a single rule
- **THEN** system applies the rule to all branches matching any of the patterns

### Requirement: Define exempted users and groups
The system SHALL allow specification of users and groups exempted from branch permission restrictions.

#### Scenario: Exempt specific users from restriction
- **WHEN** user lists specific usernames as exceptions in a branch rule
- **THEN** system allows those users to perform restricted actions on matching branches

#### Scenario: Exempt groups from restriction
- **WHEN** user lists group names as exceptions in a branch rule
- **THEN** system allows all members of those groups to perform restricted actions

#### Scenario: No exemptions specified
- **WHEN** user does not specify exemptions in a branch rule
- **THEN** system restricts the action for all users without exception

### Requirement: Support multiple restriction types
The system SHALL support all Bitbucket branch permission restriction types.

#### Scenario: Prevent branch deletion
- **WHEN** user configures a rule with "no-deletes" restriction type
- **THEN** system prevents deletion of matching branches except by exempted users

#### Scenario: Prevent rewriting history
- **WHEN** user configures a rule with "fast-forward-only" restriction type
- **THEN** system prevents force pushes and history rewrites on matching branches

#### Scenario: Prevent all writes
- **WHEN** user configures a rule with "read-only" restriction type
- **THEN** system prevents all write operations on matching branches except by exempted users

#### Scenario: Require pull request
- **WHEN** user configures a rule with "pull-request-only" restriction type
- **THEN** system prevents direct pushes and requires all changes through pull requests

### Requirement: Manage rule priority and ordering
The system SHALL apply branch permission rules in a consistent and predictable order.

#### Scenario: Multiple rules apply to same branch
- **WHEN** multiple rules match the same branch name
- **THEN** system applies all applicable rules with the most restrictive settings taking precedence

#### Scenario: Specific rule overrides wildcard rule
- **WHEN** both a wildcard rule and an exact match rule apply to a branch
- **THEN** system prioritizes the exact match rule configuration

### Requirement: Support idempotent rule management
The system SHALL support idempotent branch permission operations where applying the same configuration produces the same result.

#### Scenario: Reapply unchanged branch rules
- **WHEN** user applies the same branch permission configuration
- **THEN** system detects rules are already correct and makes no changes

#### Scenario: Update existing branch rules
- **WHEN** user modifies branch rules in YAML configuration
- **THEN** system updates the rules to match the new configuration

#### Scenario: Remove branch rules
- **WHEN** user removes branch rules from YAML configuration
- **THEN** system deletes those rules from the project

### Requirement: Validate branch rule configuration
The system SHALL validate branch permission rule syntax and values before applying changes.

#### Scenario: Valid branch rule configuration
- **WHEN** user provides well-formed branch rule configuration
- **THEN** system passes validation and proceeds to apply rules

#### Scenario: Invalid restriction type
- **WHEN** user specifies an invalid restriction type not supported by Bitbucket
- **THEN** system fails validation and lists valid restriction types

#### Scenario: Invalid branch pattern syntax
- **WHEN** user provides malformed branch pattern
- **THEN** system fails validation with error explaining pattern syntax requirements

### Requirement: Report branch rule changes
The system SHALL report all branch permission rule changes made during configuration application.

#### Scenario: New rules created
- **WHEN** system creates new branch permission rules
- **THEN** system reports each rule created with branch pattern and restrictions

#### Scenario: Rules modified
- **WHEN** system modifies existing branch permission rules
- **THEN** system reports each rule updated with old and new settings

#### Scenario: Rules deleted
- **WHEN** system removes branch permission rules
- **THEN** system reports each rule deleted with its branch pattern

### Requirement: Handle matcher complexity
The system SHALL support advanced branch matching patterns including regex-based matchers if available in Bitbucket API.

#### Scenario: Simple glob pattern
- **WHEN** user specifies a simple glob pattern like "feature/*"
- **THEN** system creates a rule using glob matching

#### Scenario: Complex branch naming pattern
- **WHEN** user specifies patterns for complex branch naming conventions
- **THEN** system correctly matches branches according to the pattern

#### Scenario: Case-sensitive branch matching
- **WHEN** user specifies branch patterns
- **THEN** system applies case-sensitive matching consistent with Git branch naming
