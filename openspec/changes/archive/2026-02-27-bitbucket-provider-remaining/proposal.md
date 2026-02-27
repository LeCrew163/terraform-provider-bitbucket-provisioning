## Why

The Bitbucket DC Terraform provider has full resource coverage but is missing plural list data sources (needed for enumeration and module reuse), generated documentation (required for registry publishing), and a CI/CD pipeline (needed for automated testing and releases). These are the final pieces to make the provider production-ready.

## What Changes

- Add `bitbucketdc_projects` data source — list all accessible projects with optional name filter
- Add `bitbucketdc_repositories` data source — list all repositories in a project with optional filter
- Generate provider documentation using `tfplugindocs` into `docs/` directory
- Add Jenkinsfile for CI/CD: lint, unit test, acceptance test (parameterised), and release stages

## Capabilities

### New Capabilities
- `data-source-projects`: List all Bitbucket DC projects, optionally filtered by name
- `data-source-repositories`: List all repositories within a project, optionally filtered by name
- `provider-documentation`: Auto-generated docs from schema via tfplugindocs
- `cicd-pipeline`: Jenkinsfile defining build, test, and release pipeline

### Modified Capabilities

## Impact

- New files: `internal/provider/data_source_projects.go`, `data_source_repositories.go`, their tests, `Jenkinsfile`, `docs/` directory
- `provider.go` updated to register two new data sources
- `tests/terraform/main.tf` updated with list data source examples
