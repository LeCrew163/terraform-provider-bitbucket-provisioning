## Why

Manual Bitbucket Data Center project provisioning is time-consuming, error-prone, and leads to inconsistent configurations across projects. DevOps/Platform Engineers currently click through the Bitbucket UI for every new project, risking missed steps and configuration drift. An automated, declarative approach using YAML configuration files will ensure consistent, auditable, and repeatable project provisioning.

## What Changes

- Introduce a Python-based CLI tool for Bitbucket Data Center project provisioning
- Generate Bitbucket API client from official OpenAPI specification (v3, Bitbucket Data Center)
- Enable declarative project configuration via YAML files
- Support exporting existing project configurations to YAML for migration
- Support CI/CD pipeline integration for automated provisioning
- Provide project structure with modular architecture for extensibility

## Capabilities

### New Capabilities
- `project-creation`: Create Bitbucket projects with standard configuration from YAML definitions
- `project-permissions`: Manage user and group permissions at project level
- `branch-permissions`: Configure branch permission rules and restrictions
- `access-keys`: Manage SSH access keys for projects
- `branch-workflow`: Configure branching model and workflow settings
- `hooks-configuration`: Set up and configure project-level hooks
- `default-reviewers`: Define and manage default reviewer rules
- `project-export`: Export existing project configurations to YAML format for migration

### Modified Capabilities
<!-- No existing capabilities to modify - this is a greenfield project -->

## Impact

**New Components:**
- Python CLI application with command structure
- OpenAPI client generation pipeline (from Bitbucket Data Center swagger spec)
- YAML schema validation for project configurations
- Configuration management module
- API integration layer

**Dependencies:**
- Bitbucket Data Center API (v3)
- OpenAPI spec: `https://dac-static.atlassian.com/server/bitbucket/10.0.swagger.v3.json`
- Python OpenAPI client generator (openapi-generator or similar)
- YAML parsing and validation library
- HTTP client library

**Affected Systems:**
- Bitbucket Data Center instance(s)
- CI/CD pipelines (for integration)

**Scope Considerations:**
- **Phase 1 (this change)**: Project provisioning (apply and validate)
- **Phase 1.5 (migration support)**: Project export for migrating existing projects to YAML
- **Phase 2 (future)**: Repository management capabilities
