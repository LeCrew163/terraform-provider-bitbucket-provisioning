# Key Decisions Summary

This document summarizes the key architectural and implementation decisions for the Terraform Provider for Bitbucket Data Center.

## Distribution and Registry

### Private Registry (Terraform Cloud/Enterprise)

**Decision:** Use Terraform Cloud/Enterprise private registry for provider distribution.

**Configuration:**
```hcl
terraform {
  required_providers {
    bitbucketdc = {
      source  = "app.terraform.io/your-org/bitbucketdc"
      version = "~> 1.0"
    }
  }
}
```

**Benefits:**
- Private provider (not publicly accessible)
- Access control via organization membership
- Internal use only (company-specific configurations)
- Version management and documentation hosting
- Seamless integration with existing Terraform Cloud workflows

**Alternative:** Public registry (`registry.terraform.io`) is available if the provider becomes open-source in the future.

## Phase 1 Implementation Scope

### 1. Clustering Support

**Decision:** Assume Bitbucket Data Center clustering is transparent to the provider.

**Rationale:**
- Bitbucket's REST API behaves identically whether single-node or clustered
- Provider connects to load balancer URL
- No special handling for routing or rate limiting
- Monitor during testing; add special handling in Phase 2 only if needed

**Implementation:**
- Single `base_url` configuration (points to load balancer)
- Standard HTTP client with connection pooling
- No cluster-aware logic required

### 2. Import Tooling

**Decision:** Standard `terraform import` in Phase 1, bulk import tooling in Phase 2.

**Phase 1 (MVP):**
```bash
# Import individual resources
terraform import bitbucketdc_project.example MYPROJ
terraform import bitbucketdc_repository.example MYPROJ/repo-slug
terraform import bitbucketdc_project_permissions.example MYPROJ
```

**Phase 2 (Enhancement):**
- Provider subcommand for bulk import
- Scans Bitbucket and generates Terraform HCL
- Example: `terraform-provider-bitbucketdc generate --project MYPROJ > generated.tf`
- Users review and adjust generated configuration before importing

**Rationale:**
- Manual import is sufficient for initial adoption
- Bulk tooling valuable for large-scale migrations but not blocking
- Can be added incrementally without affecting Phase 1 users

### 3. Custom Validation

**Decision:** Comprehensive custom validation in Phase 1.

**Validators to Implement:**

1. **Project Key Validator**
   - Pattern: `^[A-Z][A-Z0-9_]*$`
   - Uppercase alphanumeric with underscores
   - Must start with letter

2. **Repository Slug Validator**
   - Pattern: `^[a-z0-9][a-z0-9-]*[a-z0-9]$`
   - Lowercase alphanumeric with hyphens
   - Cannot start or end with hyphen

3. **Branch Pattern Validator**
   - Validate Git ref patterns (e.g., `refs/heads/*`)
   - Regex pattern validation

4. **SSH Key Format Validator**
   - Validate key type (RSA, ED25519, ECDSA)
   - Validate key structure and format

5. **Permission Level Validator**
   - Enum validation: `PROJECT_READ`, `PROJECT_WRITE`, `PROJECT_ADMIN`

**Benefits:**
- Immediate feedback before API calls
- Clear, actionable error messages
- Better user experience
- Reduces unnecessary API calls

**Implementation:**
```go
// Example validator
validators.ProjectKeyValidator() // Returns validator.String

// Usage in schema
Validators: []validator.String{
    validators.ProjectKeyValidator(),
}
```

## Technology Stack

### Core Technologies

- **Language:** Go 1.21+
- **Framework:** Terraform Plugin Framework (Protocol 6)
- **API Client:** Generated from OpenAPI spec using openapi-generator
- **Testing:** terraform-plugin-testing framework
- **Release:** GoReleaser

### Why These Choices?

**Go:**
- Standard language for Terraform providers
- Excellent tooling and ecosystem
- Strong typing and performance

**Plugin Framework (not SDKv2):**
- Modern, type-safe approach
- Better support for nested objects
- Future-proof (HashiCorp's recommended path)
- Protocol 6 supports Terraform 1.0+

**OpenAPI Generation:**
- Bitbucket provides official OpenAPI spec
- Reduces manual API client maintenance
- Type-safe client code
- Easy to regenerate when API changes

## Resource Design Philosophy

### Fine-Grained Resources

**Decision:** Use fine-grained resources rather than monolithic project resource.

**Example:**
```hcl
# Separate resources for each concern
resource "bitbucketdc_project" "example" { }
resource "bitbucketdc_project_permissions" "example" { }
resource "bitbucketdc_branch_permissions" "example" { }
resource "bitbucketdc_repository" "example" { }
```

**Benefits:**
- Single responsibility per resource
- Smaller state blast radius
- Better dependency management
- More flexible composition
- Easier testing

**Trade-off:**
- More resources to manage vs monolithic approach
- More explicit configuration required

**Rationale:** Aligns with Terraform best practices and provides better long-term maintainability.

## Authentication

### Supported Methods

1. **Personal Access Token (Recommended)**
   ```hcl
   provider "bitbucketdc" {
     base_url = "https://bitbucket.example.com"
     token    = var.bitbucket_token
   }
   ```

2. **HTTP Basic Auth**
   ```hcl
   provider "bitbucketdc" {
     base_url = "https://bitbucket.example.com"
     username = var.bitbucket_username
     password = var.bitbucket_password
   }
   ```

**Environment Variables:**
- `BITBUCKET_BASE_URL`
- `BITBUCKET_TOKEN`
- `BITBUCKET_USERNAME`
- `BITBUCKET_PASSWORD`

**Security:**
- All credentials marked as sensitive
- Never logged or exposed in plan output
- Recommend using Terraform variables with proper secret management

## Testing Strategy

### Three-Layer Testing Approach

1. **Unit Tests**
   - Schema validation
   - Validator logic
   - Model transformations
   - Import ID parsing
   - Fast, no external dependencies

2. **Integration Tests**
   - API client wrapper operations
   - Error handling and retries
   - Authentication flows
   - Mock Bitbucket API using httptest

3. **Acceptance Tests**
   - Full resource lifecycle (CRUD)
   - Import functionality
   - State refresh and drift detection
   - Multi-resource scenarios
   - Requires real Bitbucket Data Center instance

**Target:** >80% code coverage for non-acceptance tests

## Documentation

### Auto-Generated with tfplugindocs

**Structure:**
```
templates/
├── index.md.tmpl           # Provider documentation
├── resources/
│   ├── project.md.tmpl
│   └── *.md.tmpl
└── data-sources/
    └── *.md.tmpl

examples/
├── provider/
│   └── provider.tf
└── resources/
    └── bitbucketdc_*/
        └── resource.tf
```

**Benefits:**
- Documentation stays in sync with schema
- Examples are validated
- Consistent format
- Terraform Registry/Cloud compatible

## Implementation Phases

### Phase 1: Core Provider (MVP)

**Resources:**
- `bitbucketdc_project`
- `bitbucketdc_repository`
- `bitbucketdc_project_permissions`
- `bitbucketdc_branch_permissions`
- `bitbucketdc_project_access_keys`

**Data Sources:**
- `bitbucketdc_project`
- `bitbucketdc_repository`
- `bitbucketdc_user`
- `bitbucketdc_group`

**Features:**
- Full CRUD operations
- Import support
- Comprehensive validation
- Complete documentation

### Phase 2: Advanced Features

**Resources:**
- `bitbucketdc_project_hooks`
- `bitbucketdc_project_webhooks`
- `bitbucketdc_default_reviewers`
- `bitbucketdc_branch_workflow`
- `bitbucketdc_repository_permissions`
- `bitbucketdc_repository_hooks`

**Tooling:**
- Bulk import subcommand
- Configuration generator

### Phase 3: Additional Features

**As Needed:**
- Pull request settings
- Repository mirroring
- Personal repositories
- Advanced hooks configuration

## Next Steps

1. Review and approve this proposal
2. Set up Go project structure
3. Generate OpenAPI client
4. Implement provider core and authentication
5. Implement Phase 1 resources in order:
   - Project (foundational)
   - Repository (depends on project)
   - Permissions (depends on project)
   - Branch permissions (depends on project)
   - Access keys (depends on project)
6. Write comprehensive tests
7. Generate documentation
8. Release v0.1.0
9. Publish to Terraform Cloud private registry

## Questions or Concerns?

All major architectural questions have been resolved. Implementation can proceed with Phase 1 scope.
