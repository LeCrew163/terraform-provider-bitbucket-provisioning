## Requirements

### Requirement: List all accessible projects
The system SHALL provide a `bitbucketdc_projects` data source that returns all Bitbucket DC projects visible to the authenticated user.

#### Scenario: Read all projects
- **WHEN** user declares `data "bitbucketdc_projects" "all" {}`
- **THEN** system returns a list of all projects with their key, name, description, public flag, and id

#### Scenario: Filter by name substring
- **WHEN** user sets `filter = "platform"`
- **THEN** system returns only projects whose name contains "platform" (case-insensitive)

#### Scenario: No projects match filter
- **WHEN** user sets `filter = "nonexistent-xyz"`
- **THEN** system returns an empty list without error

#### Scenario: Multiple pages of results
- **WHEN** the Bitbucket instance has more projects than the default page size
- **THEN** system fetches all pages and returns the complete list

### Requirement: Expose project attributes in list
The system SHALL expose the following attributes for each project in the list: `key`, `name`, `description`, `public`, `id`.

#### Scenario: Project attributes available
- **WHEN** user reads `data.bitbucketdc_projects.all.projects[0].key`
- **THEN** the attribute contains the project's key string
