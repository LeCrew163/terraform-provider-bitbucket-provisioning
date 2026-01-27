# Terraform Provider for Bitbucket Data Center

## Overview

This proposal outlines the design and implementation of a Terraform provider for Bitbucket Data Center, enabling Infrastructure-as-Code management of Bitbucket projects, repositories, permissions, and configurations.

## Why Terraform Provider vs Python CLI?

This proposal presents an alternative approach to the existing Python CLI tool proposal (`bitbucket-provisioning-init`). Here's a comparison:

### Terraform Provider Advantages

1. **Infrastructure-as-Code Integration**
   - Native integration with existing Terraform workflows
   - Manage Bitbucket alongside other infrastructure (AWS, Kubernetes, etc.)
   - Single tool for all infrastructure management

2. **State Management**
   - Built-in state tracking and drift detection
   - Terraform handles idempotency automatically
   - Collaborative state backends (S3, Terraform Cloud)

3. **Dependency Management**
   - Automatic dependency resolution between resources
   - Parallel execution where possible
   - Predictable execution order

4. **Team Collaboration**
   - Terraform Cloud/Enterprise integration
   - PR-based workflows with plan previews
   - State locking prevents concurrent modifications

5. **Ecosystem**
   - Leverage existing Terraform modules and patterns
   - Terraform Registry for distribution
   - Large community and tooling support

6. **Planning and Preview**
   - `terraform plan` shows exact changes before applying
   - Dry-run is built into normal workflow
   - Better visibility into infrastructure changes

### Python CLI Tool Advantages

1. **Simplicity**
   - Easier to learn for teams not using Terraform
   - No Terraform knowledge required
   - Straightforward YAML configuration

2. **Standalone**
   - No dependency on Terraform installation
   - Can be used independently of other IaC tools
   - Simpler deployment model

3. **Custom Workflows**
   - More flexibility in command structure
   - Custom export/import logic
   - Easier to add custom features

### Decision Factors

Choose **Terraform Provider** if:
- Your team already uses Terraform extensively
- You need to manage Bitbucket alongside other infrastructure
- You want state management and drift detection
- You need team collaboration features (Terraform Cloud/Enterprise)
- You prefer HCL over YAML

Choose **Python CLI Tool** if:
- Your team doesn't use Terraform
- You want a standalone tool without additional dependencies
- You prefer YAML configuration
- You need custom workflows not easily expressed in Terraform

## Architecture Overview

### Technology Stack

- **Language**: Go 1.21+
- **Framework**: Terraform Plugin Framework (Protocol 6)
- **API Client**: Generated from OpenAPI spec using openapi-generator
- **Testing**: terraform-plugin-testing framework
- **Release**: GoReleaser for multi-platform builds

### Project Structure

```
terraform-provider-bitbucketdc/
├── internal/
│   ├── provider/          # Terraform resources and data sources
│   ├── client/            # Bitbucket API client wrapper
│   ├── models/            # Terraform schema models
│   └── validators/        # Custom validators
├── examples/              # Example Terraform configurations
├── docs/                  # Auto-generated documentation
├── specs/                 # OpenAPI specification
└── main.go                # Provider entrypoint
```

## Core Resources

### Resources (Manage Infrastructure)

1. **bitbucketdc_project** - Bitbucket projects
2. **bitbucketdc_repository** - Git repositories
3. **bitbucketdc_project_permissions** - Project-level permissions
4. **bitbucketdc_branch_permissions** - Branch permission rules
5. **bitbucketdc_project_access_keys** - SSH access keys

### Data Sources (Query Existing Infrastructure)

1. **bitbucketdc_project** - Query existing projects
2. **bitbucketdc_repository** - Query existing repositories
3. **bitbucketdc_user** - Query Bitbucket users
4. **bitbucketdc_group** - Query Bitbucket groups

## Example Usage

### Provider Configuration

```hcl
terraform {
  required_providers {
    bitbucketdc = {
      source  = "app.terraform.io/your-org/bitbucketdc"  # Private registry
      version = "~> 1.0"
    }
  }
}

provider "bitbucketdc" {
  base_url = "https://bitbucket.example.com"
  token    = var.bitbucket_token  # or use username/password
}
```

### Create Project and Repository

```hcl
# Create project
resource "bitbucketdc_project" "platform" {
  key         = "PLATFORM"
  name        = "Platform Engineering"
  description = "Infrastructure and platform services"
  visibility  = "private"
}

# Create repository
resource "bitbucketdc_repository" "terraform_modules" {
  project_key = bitbucketdc_project.platform.key
  slug        = "terraform-modules"
  name        = "Terraform Modules"
  description = "Shared Terraform modules"
}

# Manage permissions
resource "bitbucketdc_project_permissions" "platform" {
  project_key = bitbucketdc_project.platform.key

  user {
    name       = "admin"
    permission = "PROJECT_ADMIN"
  }

  group {
    name       = "platform-team"
    permission = "PROJECT_WRITE"
  }

  group {
    name       = "developers"
    permission = "PROJECT_READ"
  }
}

# Configure branch permissions
resource "bitbucketdc_branch_permissions" "platform" {
  project_key = bitbucketdc_project.platform.key

  restriction {
    type           = "fast-forward-only"
    branch_pattern = "refs/heads/main"
  }

  restriction {
    type           = "pull-request-only"
    branch_pattern = "refs/heads/main"
    exempted_users = ["admin"]
  }
}
```

### Import Existing Infrastructure

```bash
# Import existing project
terraform import bitbucketdc_project.platform PLATFORM

# Import existing repository
terraform import bitbucketdc_repository.terraform_modules PLATFORM/terraform-modules

# Import permissions
terraform import bitbucketdc_project_permissions.platform PLATFORM
```

### Query Existing Resources

```hcl
# Query existing project
data "bitbucketdc_project" "legacy" {
  key = "LEGACY"
}

# Use in new repository
resource "bitbucketdc_repository" "new_service" {
  project_key = data.bitbucketdc_project.legacy.key
  slug        = "new-service"
  name        = "New Service"
}
```

## Key Features

### 1. Full Resource Lifecycle Management

- **Create**: Provision new Bitbucket resources
- **Read**: Query and refresh state
- **Update**: Modify existing resources
- **Delete**: Clean up resources
- **Import**: Adopt existing resources into Terraform management

### 2. State Management

- Terraform state tracks all managed resources
- Drift detection shows manual changes
- State backends enable team collaboration
- State locking prevents concurrent modifications

### 3. Plan and Apply Workflow

```bash
# Preview changes
terraform plan

# Apply changes
terraform apply

# Destroy resources
terraform destroy
```

### 4. Dependencies

Terraform automatically handles dependencies:

```hcl
# Repository depends on project (automatic)
resource "bitbucketdc_repository" "repo" {
  project_key = bitbucketdc_project.proj.key  # Dependency
  # ...
}

# Explicit dependencies if needed
resource "bitbucketdc_project_permissions" "perms" {
  project_key = bitbucketdc_project.proj.key
  depends_on  = [bitbucketdc_repository.repo]
}
```

### 5. Modules

Create reusable modules:

```hcl
# modules/bitbucket-project/main.tf
variable "project_key" {}
variable "project_name" {}
variable "admin_group" {}

resource "bitbucketdc_project" "this" {
  key  = var.project_key
  name = var.project_name
}

resource "bitbucketdc_project_permissions" "this" {
  project_key = bitbucketdc_project.this.key

  group {
    name       = var.admin_group
    permission = "PROJECT_ADMIN"
  }
}

# Use the module
module "platform_project" {
  source = "./modules/bitbucket-project"

  project_key  = "PLATFORM"
  project_name = "Platform Engineering"
  admin_group  = "platform-admins"
}
```

## Implementation Phases

### Phase 1: Core Provider (MVP)

- [x] Provider configuration and authentication
- [x] Project resource
- [x] Repository resource
- [x] Project permissions resource
- [x] Branch permissions resource
- [x] Project access keys resource
- [x] Basic data sources (project, repository, user, group)
- [x] Import functionality
- [x] Documentation and examples

### Phase 2: Advanced Features

- [ ] Project hooks resource
- [ ] Project webhooks resource
- [ ] Default reviewers resource
- [ ] Branch workflow resource
- [ ] Repository permissions resource
- [ ] Repository hooks resource
- [ ] Additional data sources (list repositories, list projects)

### Phase 3: Additional Features

- [ ] Pull request settings resource
- [ ] Repository mirroring configuration
- [ ] Personal repository support
- [ ] Advanced validation and testing

## Testing Strategy

### Unit Tests
- Schema validation
- Model transformations
- Validator logic
- Import ID parsing

### Integration Tests
- API client operations
- Error handling
- Authentication flows

### Acceptance Tests
- Full resource lifecycle (CRUD)
- Import functionality
- State refresh and drift detection
- Multi-resource scenarios

## Documentation

### Auto-Generated Documentation

Using `tfplugindocs`:
- Provider configuration
- Resource schemas with examples
- Data source schemas with examples
- Import instructions

### Additional Guides

- Getting started guide
- Authentication setup
- Migration from manual management
- Best practices
- Troubleshooting

## Release and Distribution

### GitHub Releases

Using GoReleaser:
- Multi-platform builds (Linux, macOS, Windows)
- Signed releases (GPG)
- Automated changelog
- Release artifacts

### Terraform Private Registry

**Distribution Strategy:**
- Host provider in Terraform Cloud/Enterprise private registry
- Private distribution for internal use (not public open-source)
- Access control via Terraform Cloud organization membership
- Provider namespace: `app.terraform.io/<your-org>/bitbucketdc`

**Benefits:**
- Private provider (not publicly accessible)
- Version management and release tracking
- Provider documentation hosting
- Seamless `terraform init` experience
- Integration with Terraform Cloud features

**Alternative: Public Terraform Registry**
If making the provider open-source in the future:
- Submit to `registry.terraform.io`
- Public namespace: `registry.terraform.io/<namespace>/bitbucketdc`
- Requires HashiCorp review and approval
- Provider becomes publicly available

## Migration Path

### From Manual Management

1. Write Terraform configuration for existing resources
2. Import resources into Terraform state
3. Run `terraform plan` to verify configuration matches reality
4. Begin managing changes through Terraform

### From Python CLI Tool

If migrating from the Python CLI tool proposal:

1. Convert YAML configurations to HCL
2. Import resources into Terraform state
3. Use Terraform for future changes

## Files in This Proposal

- **proposal.md** - High-level overview and impact
- **design.md** - Detailed architectural decisions
- **tasks.md** - Implementation task breakdown
- **specs/** - Detailed specifications for each resource
  - project-resource/spec.md
  - repository-resource/spec.md
  - project-permissions-resource/spec.md
  - branch-permissions-resource/spec.md
- **README.md** - This file

## Next Steps

1. Review and approve this proposal
2. Set up project structure and dependencies
3. Generate OpenAPI client
4. Implement core provider and resources
5. Write comprehensive tests
6. Generate documentation
7. Release v0.1.0
8. Submit to Terraform Registry

## Questions?

Please review the detailed design document (`design.md`) for architectural decisions and trade-offs. Each resource has a detailed specification in the `specs/` directory.
