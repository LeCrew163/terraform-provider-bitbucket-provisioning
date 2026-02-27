## Requirements

### Requirement: List all repositories in a project
The system SHALL provide a `bitbucketdc_repositories` data source that returns all repositories within a given Bitbucket DC project.

#### Scenario: Read all repositories in a project
- **WHEN** user declares `data "bitbucketdc_repositories" "all" { project_key = "MYPROJECT" }`
- **THEN** system returns a list of all repositories with their slug, name, description, public flag, state, forkable, default_branch, clone_url_http, clone_url_ssh, and id

#### Scenario: Filter by name substring
- **WHEN** user sets `filter = "service"`
- **THEN** system returns only repositories whose name contains "service" (case-insensitive)

#### Scenario: No repositories match filter
- **WHEN** user sets `filter = "nonexistent-xyz"`
- **THEN** system returns an empty list without error

#### Scenario: Project not found
- **WHEN** user specifies a `project_key` that does not exist
- **THEN** system returns an error indicating the project was not found

#### Scenario: Multiple pages of results
- **WHEN** the project has more repositories than the default page size
- **THEN** system fetches all pages and returns the complete list

### Requirement: Expose repository attributes in list
The system SHALL expose the following attributes for each repository: `slug`, `name`, `description`, `public`, `state`, `forkable`, `default_branch`, `clone_url_http`, `clone_url_ssh`, `id` (`{project_key}/{slug}`).

#### Scenario: Repository attributes available
- **WHEN** user reads `data.bitbucketdc_repositories.all.repositories[0].slug`
- **THEN** the attribute contains the repository's slug string
