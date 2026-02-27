## Why

Manual Bitbucket Data Center project and repository management is time-consuming, error-prone, and leads to inconsistent configurations. Infrastructure teams increasingly adopt Infrastructure-as-Code (IaC) practices with Terraform, but **there is currently NO Terraform provider for Bitbucket Data Center**.

The only existing Bitbucket provider ([DrFaust92/terraform-provider-bitbucket](https://github.com/DrFaust92/terraform-provider-bitbucket)) is specifically designed for **Bitbucket Cloud** and is incompatible with Data Center due to fundamental API and architecture differences. This forces teams using Bitbucket Data Center to rely on external provisioning scripts or manual processes.

A native Terraform provider for Bitbucket Data Center will enable declarative, version-controlled, and auditable infrastructure management integrated seamlessly with existing Terraform workflows, filling a genuine gap in the Terraform ecosystem.

## What Changes

- Develop a Terraform provider for Bitbucket Data Center (not Cloud)
- Support full resource lifecycle management (Create, Read, Update, Delete)
- Generate Bitbucket API client from official OpenAPI specification (v3)
- Enable declarative infrastructure definition using HCL (Terraform's configuration language)
- Provide import functionality for existing Bitbucket resources
- Integrate with Terraform state management for drift detection
- Support Terraform Cloud/Enterprise for team collaboration

## Capabilities

### New Capabilities
- `project-resource`: Terraform resource for managing Bitbucket projects
- `project-permissions-resource`: Manage user and group permissions at project level
- `branch-permissions-resource`: Configure branch permission rules and restrictions
- `project-access-keys-resource`: Manage SSH access keys for projects
- `branch-workflow-resource`: Configure branching model and workflow settings
- `project-hooks-resource`: Set up and configure project-level hooks
- `default-reviewers-resource`: Define and manage default reviewer rules
- `repository-resource`: Manage repositories within projects
- `repository-permissions-resource`: Manage repository-level permissions
- `repository-hooks-resource`: Configure repository-level hooks
- `project-datasource`: Data source for querying existing projects
- `repository-datasource`: Data source for querying existing repositories
- `user-datasource`: Data source for querying Bitbucket users
- `group-datasource`: Data source for querying Bitbucket groups

### Modified Capabilities
<!-- No existing capabilities to modify - this is a new provider -->

## Impact

**New Components:**
- Terraform provider plugin written in Go
- Bitbucket Data Center API client (generated from OpenAPI spec)
- Provider configuration and authentication module
- Resource implementations for each Bitbucket entity
- Data source implementations for read-only queries
- Import functionality for existing resources
- Comprehensive test suite (unit, integration, acceptance tests)
- Provider documentation following Terraform Registry standards

**Dependencies:**
- Bitbucket Data Center API (v3)
- OpenAPI spec: `https://dac-static.atlassian.com/server/bitbucket/10.0.swagger.v3.json`
- Terraform Plugin Framework (v1.x)
- Go OpenAPI client generator (go-swagger or openapi-generator)
- HashiCorp plugin testing framework

**Affected Systems:**
- Bitbucket Data Center instance(s)
- Terraform workflows (local, Terraform Cloud, Terraform Enterprise)
- CI/CD pipelines using Terraform

**Scope Considerations:**
- **Phase 1 (this change)**: Core provider with project and basic repository management
  - Projects, permissions, branch permissions, access keys
  - Basic repository management
  - Provider configuration and authentication
- **Phase 2 (future)**: Advanced features
  - Webhooks and advanced hooks configuration
  - Branch workflow and default reviewers
  - Repository-level fine-grained permissions
  - Pull request settings
- **Phase 3 (future)**: Additional features
  - Bitbucket Data Center settings
  - Personal repositories
  - Cluster configuration (if applicable)

**Distribution:**
- Published to **Terraform Cloud/Enterprise private registry** (internal use)
- Provider name: `terraform-provider-bitbucketdc`
- Namespace: `app.terraform.io/<your-org>/bitbucketdc` (private)
- Documentation hosted in private registry
- Semantic versioning for releases
- Access controlled via organization membership
- Note: Can be published to public Terraform Registry later if open-sourced
