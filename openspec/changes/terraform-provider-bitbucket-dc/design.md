## Context

This is a greenfield Go project to create a Terraform provider for Bitbucket Data Center. Infrastructure teams need to manage Bitbucket resources (projects, repositories, permissions) as code alongside their other infrastructure. A native Terraform provider enables declarative management, state tracking, drift detection, and integration with existing Terraform workflows.

**Constraints:**
- Must work with Bitbucket Data Center (not Cloud)
- Must follow Terraform provider best practices and conventions
- Must support Terraform 1.0+ (using Protocol version 6)
- Must be maintainable by teams familiar with Go and Terraform
- API client must stay synchronized with Bitbucket API changes
- Must pass HashiCorp provider quality standards for potential registry listing

**Stakeholders:**
- Primary users: DevOps/Platform Engineers, SREs using Terraform
- Secondary users: Developers defining infrastructure in Terraform modules
- Terraform Cloud/Enterprise users requiring team collaboration

## Goals / Non-Goals

**Goals:**
- Create production-ready Terraform provider for Bitbucket Data Center
- Define optimal project structure following Terraform provider conventions
- Establish OpenAPI client generation workflow for Go
- Design resource and data source architecture covering core use cases
- Enable full Terraform lifecycle support (CRUD + Import)
- Support state management and drift detection
- Provide comprehensive documentation and examples

**Non-Goals:**
- Bitbucket Cloud support (different API)
- Management of Bitbucket Data Center server configuration
- Migration tools from other providers
- Backup/restore functionality
- Management of Bitbucket system settings

## Decisions

### 1. Project Structure: Standard Terraform Provider Layout

**Decision:** Use standard Terraform provider project structure with internal package organization.

**Rationale:**
- Follows HashiCorp conventions and community standards
- Clear separation between provider framework code and business logic
- Easier onboarding for contributors familiar with Terraform providers
- Better tooling support (IDE, linters, generators)

**Structure:**
```
terraform-provider-bitbucketdc/
├── internal/
│   ├── provider/          # Provider implementation
│   │   ├── provider.go           # Main provider definition
│   │   ├── provider_test.go      # Provider tests
│   │   ├── resource_project.go   # Project resource
│   │   ├── resource_repository.go
│   │   ├── resource_*.go         # Other resources
│   │   ├── data_source_project.go
│   │   └── data_source_*.go      # Data sources
│   ├── client/            # Bitbucket API client wrapper
│   │   ├── client.go             # Client initialization
│   │   ├── projects.go           # Project operations
│   │   ├── repositories.go       # Repository operations
│   │   └── generated/            # Generated OpenAPI client
│   ├── models/            # Terraform schema models
│   └── validators/        # Custom validators
├── examples/              # Example Terraform configurations
├── docs/                  # Auto-generated documentation
├── templates/             # Documentation templates
├── tools/                 # Code generation tools
├── .goreleaser.yml        # Release configuration
├── main.go                # Provider entrypoint
├── go.mod
├── go.sum
└── README.md
```

**Alternatives Considered:**
- Monolithic single-file provider: Rejected - not maintainable for complex providers
- Flat internal structure: Rejected - harder to organize as provider grows

### 2. Terraform Framework: Terraform Plugin Framework

**Decision:** Use Terraform Plugin Framework (protocol version 6) instead of SDKv2.

**Rationale:**
- Modern framework with better type safety
- Native support for nested objects and collections
- Better validation and diagnostics capabilities
- Built-in support for private state
- Better performance for large configurations
- Future-proof (HashiCorp's recommended path forward)
- Protocol version 6 supports Terraform 1.0+

**Implementation:**
```go
import (
    "github.com/hashicorp/terraform-plugin-framework/provider"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/datasource"
)
```

**Alternatives Considered:**
- SDKv2: Rejected - older framework, less type-safe, more maintenance burden
- Terraform Plugin Go: Rejected - too low-level, more implementation work

### 3. OpenAPI Client Generation: openapi-generator (Go)

**Decision:** Use `openapi-generator` to generate Go client from Bitbucket OpenAPI spec.

**Rationale:**
- Mature, widely-used tool with good Go support
- Generates idiomatic Go code with proper types
- Supports OpenAPI 3.0 specification
- Active community and maintenance
- Can customize generation with templates

**Implementation:**
```bash
# Makefile target
generate-client:
    openapi-generator-cli generate \
        -i specs/bitbucket-openapi.json \
        -g go \
        -o internal/client/generated \
        --additional-properties=packageName=bitbucket
```

**Wrapper Layer:**
- Thin wrapper in `internal/client/` for:
  - Authentication configuration (Personal Access Tokens, HTTP Basic)
  - Error translation to Terraform diagnostics
  - Retry logic with exponential backoff
  - Base URL management
  - Context propagation

**Alternatives Considered:**
- `go-swagger`: Rejected - more complex, less idiomatic Go output
- Manual client: Rejected - too much maintenance burden
- `oapi-codegen`: Considered but less mature than openapi-generator

### 4. Provider Configuration: Token-Based Authentication

**Decision:** Support Personal Access Tokens and HTTP Basic Auth with flexible configuration.

**Rationale:**
- Data Center supports both PAT and HTTP Basic Auth
- PATs are more secure (scoped, revocable)
- Flexible configuration matches Terraform patterns
- Environment variable support for CI/CD

**Configuration Schema:**
```hcl
provider "bitbucketdc" {
  base_url = "https://bitbucket.example.com"  # Required

  # Option 1: Personal Access Token (recommended)
  token = var.bitbucket_token

  # Option 2: HTTP Basic Auth
  username = var.bitbucket_username
  password = var.bitbucket_password

  # Optional settings
  insecure_skip_verify = false  # Skip TLS verification (not recommended)
  timeout              = 30      # Request timeout in seconds
}
```

**Environment Variables:**
- `BITBUCKET_BASE_URL`
- `BITBUCKET_TOKEN`
- `BITBUCKET_USERNAME`
- `BITBUCKET_PASSWORD`

**Alternatives Considered:**
- OAuth: Rejected - unnecessary complexity for server-to-server
- SSH keys: Rejected - not supported by REST API
- Multiple instance support: Deferred to future enhancement (use provider aliases)

### 5. Resource Design: Fine-Grained Resources

**Decision:** Design fine-grained resources rather than monolithic project resource.

**Rationale:**
- Follows Terraform best practices (single responsibility)
- Better state management (smaller blast radius for changes)
- More flexible composition in modules
- Clearer dependencies between resources
- Easier to test individual resources

**Resource Structure:**
```hcl
# Project resource
resource "bitbucketdc_project" "example" {
  key         = "MYPROJ"
  name        = "My Project"
  description = "Example project"
  visibility  = "private"
}

# Project permissions (separate resource)
resource "bitbucketdc_project_permissions" "example" {
  project_key = bitbucketdc_project.example.key

  user {
    name       = "john.doe"
    permission = "PROJECT_ADMIN"
  }

  group {
    name       = "developers"
    permission = "PROJECT_WRITE"
  }
}

# Branch permissions
resource "bitbucketdc_branch_permissions" "example" {
  project_key = bitbucketdc_project.example.key

  restriction {
    type         = "fast-forward-only"
    branch_pattern = "refs/heads/main"
  }
}
```

**Alternatives Considered:**
- Monolithic project resource: Rejected - harder to manage, forces unnecessary updates
- Nested resources: Rejected - complicates state management and imports

### 6. State Management: Normalized Attributes

**Decision:** Store normalized, API-compatible values in Terraform state.

**Rationale:**
- Reduces state drift
- Simplifies refresh operations
- Clear mapping between state and API responses
- Easier debugging and troubleshooting

**Implementation:**
- Use Terraform types that match API types
- Store IDs returned from API for all resources
- Normalize computed fields (timestamps, defaults)
- Use `Required`, `Optional`, `Computed` appropriately

**Example:**
```go
Schema: schema.Schema{
    Attributes: map[string]schema.Attribute{
        "id": schema.StringAttribute{
            Computed: true,
            PlanModifiers: []planmodifier.String{
                stringplanmodifier.UseStateForUnknown(),
            },
        },
        "key": schema.StringAttribute{
            Required: true,
            PlanModifiers: []planmodifier.String{
                stringplanmodifier.RequiresReplace(),
            },
        },
        "name": schema.StringAttribute{
            Required: true,
        },
    },
}
```

**Alternatives Considered:**
- Store raw API responses: Rejected - causes unnecessary diffs
- User-friendly transformations in state: Rejected - complicates refresh logic

### 7. Import Support: ID-Based Import

**Decision:** Implement import for all resources using standard ID patterns.

**Rationale:**
- Critical for adopting existing Bitbucket infrastructure
- Enables gradual migration to Terraform management
- Standard Terraform workflow

**Import ID Patterns:**
```bash
# Project
terraform import bitbucketdc_project.example PROJKEY

# Repository
terraform import bitbucketdc_repository.example PROJKEY/repo-slug

# Project permissions
terraform import bitbucketdc_project_permissions.example PROJKEY

# Branch permissions
terraform import bitbucketdc_branch_permissions.example PROJKEY/restriction-id

# User permission (specific)
terraform import bitbucketdc_project_user_permission.example PROJKEY/username
```

**Implementation:**
- Each resource implements `ImportState` method
- Parse import ID and populate resource data
- Call Read method to populate full state

**Alternatives Considered:**
- Complex import IDs with metadata: Rejected - poor user experience
- No import support: Rejected - blocks adoption

### 8. Error Handling: Rich Diagnostics

**Decision:** Provide detailed, actionable error messages using Terraform diagnostics.

**Rationale:**
- Better user experience
- Clear guidance for resolving issues
- Standard Terraform error format
- Support for warnings and multiple errors

**Implementation:**
```go
// Error with detail
resp.Diagnostics.AddError(
    "Failed to Create Project",
    fmt.Sprintf("Could not create project '%s': %s\n\n"+
        "Please verify the project key is unique and you have permissions.",
        data.Key.ValueString(), err),
)

// Warning
resp.Diagnostics.AddWarning(
    "Legacy Hook Configuration",
    "Hook 'com.example.old-hook' is deprecated. Consider migrating to 'com.example.new-hook'.",
)
```

**Error Categories:**
- Authentication errors: Clear message about credentials
- Permission errors: Explain required permissions
- Validation errors: Show invalid values and constraints
- API errors: Include HTTP status and API error details
- Rate limiting: Suggest retry or backoff

**Alternatives Considered:**
- Generic error messages: Rejected - poor user experience
- Panic on errors: Rejected - not Terraform convention

### 9. Testing Strategy: Multi-Layer Testing

**Decision:** Comprehensive testing with unit, integration, and acceptance tests.

**Test Layers:**

1. **Unit Tests**: Individual functions, schema validation
   - Test schema definitions
   - Test ID parsing for imports
   - Test model transformations
   - Fast, no external dependencies

2. **Integration Tests**: Client operations against mock server
   - Test API client wrapper
   - Test error handling
   - Mock Bitbucket API responses
   - Use `httptest` package

3. **Acceptance Tests**: Full Terraform operations against real Bitbucket
   - Test complete resource lifecycle (CRUD)
   - Test imports
   - Test updates and drift detection
   - Use `terraform-plugin-testing` framework
   - Require test Bitbucket Data Center instance

**Framework:**
```go
import (
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccProject_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccProjectConfig_basic("TEST", "Test Project"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bitbucketdc_project.test", "key", "TEST"),
                    resource.TestCheckResourceAttr("bitbucketdc_project.test", "name", "Test Project"),
                ),
            },
            // Test import
            {
                ResourceName:      "bitbucketdc_project.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}
```

**Alternatives Considered:**
- Acceptance tests only: Rejected - too slow for development
- No acceptance tests: Rejected - insufficient quality assurance

### 10. Documentation: Auto-Generated from Templates

**Decision:** Use `tfplugindocs` to generate documentation from templates and schema.

**Rationale:**
- Consistent documentation format
- Documentation stays in sync with schema
- Terraform Registry compatible
- Reduces maintenance burden
- Examples are validated

**Structure:**
```
templates/
├── index.md.tmpl           # Provider documentation
├── resources/
│   ├── project.md.tmpl
│   ├── repository.md.tmpl
│   └── *.md.tmpl
└── data-sources/
    ├── project.md.tmpl
    └── *.md.tmpl

examples/
├── provider/
│   └── provider.tf
├── resources/
│   ├── bitbucketdc_project/
│   │   └── resource.tf
│   └── bitbucketdc_repository/
│       └── resource.tf
└── data-sources/
    └── bitbucketdc_project/
        └── data-source.tf
```

**Generation:**
```bash
make docs
# Uses: tfplugindocs generate --provider-name bitbucketdc
```

**Alternatives Considered:**
- Manual documentation: Rejected - gets out of sync quickly
- In-code documentation only: Rejected - insufficient for users

### 11. Versioning and Release: GoReleaser

**Decision:** Use GoReleaser for automated, multi-platform releases.

**Rationale:**
- Standard tool for Go projects
- Builds for multiple OS/architecture combinations
- Creates GitHub releases automatically
- Signs releases (GPG)
- Generates checksums
- Terraform Registry compatible

**Configuration:**
```yaml
# .goreleaser.yml
version: 2
builds:
  - env:
      - CGO_ENABLED=0
    mod_timestamp: '{{ .CommitTimestamp }}'
    flags:
      - -trimpath
    ldflags:
      - '-s -w -X main.version={{.Version}}'
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    binary: '{{ .ProjectName }}_v{{ .Version }}'

archives:
  - format: zip
    name_template: '{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}'

checksum:
  name_template: '{{ .ProjectName }}_{{ .Version }}_SHA256SUMS'
  algorithm: sha256

signs:
  - artifacts: checksum
    args:
      - "--batch"
      - "--local-user"
      - "{{ .Env.GPG_FINGERPRINT }}"
      - "--output"
      - "${signature}"
      - "--detach-sign"
      - "${artifact}"

release:
  github:
    owner: <org>
    name: terraform-provider-bitbucketdc
```

**Release Process:**
```bash
# Tag release
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# GoReleaser builds and publishes
goreleaser release --clean
```

**Alternatives Considered:**
- Manual builds: Rejected - error-prone, inconsistent
- GitHub Actions only: Considered but GoReleaser provides more features

### 12. Idempotency: Built-In via Terraform

**Decision:** Rely on Terraform's built-in idempotency and state management.

**Rationale:**
- Terraform handles plan/apply cycles
- State tracking prevents unnecessary API calls
- Refresh operations detect drift
- Standard Terraform behavior users expect

**Implementation:**
- Implement Read to populate current state
- Implement Update to apply only changes
- Use `RequiresReplace` for immutable attributes
- Leverage Terraform's diff engine

**Example:**
```go
// Read fetches current state
func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var data projectResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    project, err := r.client.GetProject(ctx, data.Key.ValueString())
    if err != nil {
        if isNotFound(err) {
            resp.State.RemoveResource(ctx)
            return
        }
        resp.Diagnostics.AddError("Read Error", err.Error())
        return
    }

    // Update state from API
    data.Name = types.StringValue(project.Name)
    data.Description = types.StringValue(project.Description)
    // ...

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

**Alternatives Considered:**
- Custom idempotency logic: Rejected - Terraform already provides this
- Always replace resources: Rejected - wasteful and disruptive

### 13. Resource Lifecycle: Standard CRUD Pattern

**Decision:** Implement standard CRUD methods for all resources with proper error handling.

**Rationale:**
- Clear lifecycle expectations
- Standard Terraform behavior
- Proper cleanup on resource deletion
- Consistent error handling

**Resource Interface:**
```go
type Resource interface {
    Metadata(context.Context, MetadataRequest, *MetadataResponse)
    Schema(context.Context, SchemaRequest, *SchemaResponse)
    Create(context.Context, CreateRequest, *CreateResponse)
    Read(context.Context, ReadRequest, *ReadResponse)
    Update(context.Context, UpdateRequest, *UpdateResponse)
    Delete(context.Context, DeleteRequest, *DeleteResponse)
}
```

**Best Practices:**
- **Create**: Set ID after successful creation
- **Read**: Remove from state if not found (deleted outside Terraform)
- **Update**: Only call API for changed attributes
- **Delete**: Handle already-deleted resources gracefully
- All methods: Use context for cancellation, add detailed diagnostics

**Alternatives Considered:**
- Partial CRUD implementation: Rejected - incomplete provider experience
- Custom lifecycle methods: Rejected - breaks Terraform conventions

### 14. Data Sources: Read-Only Queries

**Decision:** Implement data sources for querying existing Bitbucket resources.

**Rationale:**
- Reference existing resources not managed by Terraform
- Compose infrastructure using existing projects/repos
- Enable module inputs based on current state
- Standard Terraform pattern

**Example Data Sources:**
```hcl
# Query existing project
data "bitbucketdc_project" "existing" {
  key = "LEGACY"
}

# Use in new resource
resource "bitbucketdc_repository" "new_repo" {
  project_key = data.bitbucketdc_project.existing.key
  name        = "new-repository"
  slug        = "new-repository"
}

# Query repositories
data "bitbucketdc_repositories" "all" {
  project_key = "MYPROJ"
}

# Query users
data "bitbucketdc_user" "admin" {
  username = "admin"
}
```

**Implementation:**
- Similar to Read operation
- No state modification
- Support for filtering where applicable
- Clear error messages for not found

**Alternatives Considered:**
- Resources only: Rejected - limits flexibility
- Full query DSL: Deferred - add if needed

### 15. Dependency Management: Go Modules

**Decision:** Use Go modules for dependency management with vendoring.

**Rationale:**
- Standard Go dependency management
- Reproducible builds
- Version pinning
- Vendoring ensures offline builds

**Structure:**
```
go.mod  # Dependency declarations
go.sum  # Checksums for verification
vendor/ # Vendored dependencies (committed)
```

**Key Dependencies:**
- `github.com/hashicorp/terraform-plugin-framework`
- `github.com/hashicorp/terraform-plugin-go`
- `github.com/hashicorp/terraform-plugin-testing`
- Generated OpenAPI client dependencies

**Commands:**
```bash
# Add dependency
go get github.com/example/package

# Update dependencies
go get -u ./...

# Vendor dependencies
go mod vendor

# Verify dependencies
go mod verify
```

**Alternatives Considered:**
- No vendoring: Rejected - less reproducible
- dep or other tools: Rejected - Go modules are standard

## Risks / Trade-offs

### Risk 1: OpenAPI Spec Changes

**Risk:** Bitbucket updates API, breaking generated client.

**Mitigation:**
- Pin OpenAPI spec version in repository
- Version lock generated client
- Add CI checks for spec changes
- Semantic versioning for provider releases
- Document supported Bitbucket versions

### Risk 2: Bitbucket Data Center Version Compatibility

**Risk:** Different Bitbucket versions have different API capabilities.

**Mitigation:**
- Document minimum supported Bitbucket version (10.0+)
- Add version detection in provider
- Graceful feature degradation
- Clear error messages for unsupported versions
- Test against multiple Bitbucket versions

### Risk 3: State Drift

**Risk:** Manual changes in Bitbucket UI cause state drift.

**Mitigation:**
- Terraform refresh detects drift
- Plan shows differences before apply
- Document best practices (avoid manual changes)
- Consider implementing lifecycle ignore rules where appropriate

### Risk 4: Partial Application Failures

**Risk:** Some resources created, others fail - inconsistent state.

**Mitigation:**
- Terraform handles partial failures naturally
- Failed resources stay in state for retry
- Detailed error messages indicate which resource failed
- User can apply again to continue
- Document recovery procedures

### Risk 5: Breaking Changes in Terraform Framework

**Risk:** Terraform Plugin Framework introduces breaking changes.

**Mitigation:**
- Pin framework version in go.mod
- Monitor framework releases and changelogs
- Test against new framework versions before upgrading
- Follow semantic versioning for provider

### Risk 6: Provider Performance with Large Configurations

**Risk:** Slow performance with many resources or large state files.

**Mitigation:**
- Implement efficient API client with connection pooling
- Use pagination for list operations
- Add request rate limiting and retry logic
- Profile and optimize hot paths
- Document performance considerations

### Trade-off 1: Fine-Grained vs Monolithic Resources

**Decision:** Fine-grained resources (see Decision #5).

**Trade-off:**
- Pro: Better state management, smaller blast radius, more flexible
- Con: More resources to manage, more complex configurations

**Rationale:** Fine-grained resources align with Terraform best practices and provide better long-term maintainability.

### Trade-off 2: Generated Code in Repository

**Decision:** Commit generated API client to version control.

**Trade-off:**
- Pro: Reproducible builds, no generation step for users, easier CI/CD
- Con: Larger repository, potential merge conflicts

**Rationale:** Reproducibility and ease of development outweigh repository size concerns.

### Trade-off 3: Go Language Requirement

**Decision:** Use Go as implementation language.

**Trade-off:**
- Pro: Standard for Terraform providers, good tooling, HashiCorp ecosystem
- Con: Team must know Go, different from Python proposal

**Rationale:** Go is the de facto standard for Terraform providers. Using any other language would create significant friction with the ecosystem.

### Trade-off 4: Terraform Plugin Framework vs SDKv2

**Decision:** Use Plugin Framework (see Decision #2).

**Trade-off:**
- Pro: Modern, type-safe, better features, future-proof
- Con: Less mature than SDKv2, fewer examples, potential bugs

**Rationale:** Plugin Framework is the future of Terraform provider development and provides significant quality-of-life improvements.

## Migration Plan

**N/A** - This is a greenfield provider with no existing codebase to migrate.

**Adoption Path:**
1. Publish provider to GitHub releases
2. Submit to Terraform Registry (requires HashiCorp review)
3. Document installation and usage
4. Provide import guides for existing Bitbucket infrastructure
5. Create example modules and best practices
6. Conduct team training sessions

**Import Strategy for Existing Infrastructure:**
```bash
# Step 1: Write Terraform config for existing resource
resource "bitbucketdc_project" "existing_project" {
  key         = "PROJ"
  name        = "Existing Project"
  description = "Migrated project"
  visibility  = "private"
}

# Step 2: Import existing resource
terraform import bitbucketdc_project.existing_project PROJ

# Step 3: Run plan to verify
terraform plan
```

**Rollback:**
- Provider versions can be pinned in Terraform configuration
- Previous provider versions remain available
- Changes made to Bitbucket: Use Terraform state to understand changes, manual rollback if needed
- Future: Consider implementing "destroy" protection for critical resources

## Resolved Questions

1. **Q: Should we support both Bitbucket Data Center and Cloud?**
   - **DECIDED:** Data Center only for Phase 1
   - Cloud has significantly different API structure
   - Separate providers are common (e.g., aws vs awscc)
   - Can create separate provider for Cloud later

2. **Q: Plugin Framework vs SDKv2?**
   - **DECIDED:** Plugin Framework (see Decision #2)
   - Modern, type-safe, better features
   - HashiCorp's recommended direction

3. **Q: Fine-grained resources vs monolithic?**
   - **DECIDED:** Fine-grained (see Decision #5)
   - Better state management and flexibility
   - Follows Terraform best practices

4. **Q: How to handle resource dependencies (e.g., project → repository)?**
   - **DECIDED:** Use standard Terraform dependency mechanisms
   - Resources reference each other using attributes
   - Terraform automatically determines dependency order
   - Example: `project_key = bitbucketdc_project.example.key`

5. **Q: Should we generate entire provider from OpenAPI spec?**
   - **DECIDED:** No - only generate client library
   - Provider logic requires manual implementation
   - Schema mapping and Terraform-specific behavior can't be auto-generated
   - Generated client wrapped in custom layer

6. **Q: How to handle sensitive values (tokens, webhook secrets)?**
   - **DECIDED:** Use Terraform's native sensitive value support
   - Mark schema attributes as `Sensitive: true`
   - Values hidden in plan output and logs
   - Users manage secrets via Terraform variables/vault integration

7. **Q: Support for Terraform state backends?**
   - **DECIDED:** Not provider's concern
   - Terraform handles state backend selection
   - Provider works with any backend
   - Document best practices for team collaboration

8. **Q: Handling of long-running operations?**
   - **DECIDED:** Use context with timeouts
   - Default timeouts in provider config
   - Users can override with resource-level timeouts
   - Poll for completion where necessary (e.g., hook deployment)

## Resolved Questions (Additional)

9. **Q: Namespace for Terraform Registry?**
   - **DECIDED:** Use company/organization namespace with private registry
   - **Implementation:**
     - For internal/private use: Terraform Cloud/Enterprise private registry
     - Format: `app.terraform.io/<org>/bitbucketdc` (private)
     - Public registry (`registry.terraform.io`) is for open-source only
   - **Rationale:** Company wants private provider distribution. Terraform Cloud/Enterprise provides private registry for internal providers with access controls, versioning, and team collaboration features.

10. **Q: Support for Bitbucket Data Center clustering?**
    - **DECIDED:** Assume clustering is transparent to the provider (no special handling in Phase 1)
    - **Implementation:**
      - Provider connects to load balancer URL (same as single-node)
      - Standard REST API behavior (clustering is handled by Bitbucket)
      - No special rate limiting or routing logic
      - Monitor in testing; add special handling in Phase 2 if issues arise
    - **Rationale:** Bitbucket Data Center clustering is designed to be transparent to API clients. The load balancer handles routing and the API behaves identically whether single-node or clustered. Adding special handling without proven need would add unnecessary complexity.

11. **Q: Advanced import scenarios (bulk import)?**
    - **DECIDED:** Phase 2 - implement as provider subcommand
    - **Phase 1:** Standard `terraform import` for individual resources
    - **Phase 2:** Add subcommand to scan Bitbucket and generate Terraform HCL
      - Example: `terraform-provider-bitbucketdc generate --project MYPROJ > generated.tf`
      - Generates resource blocks for all project components
      - Users can review and adjust before importing
    - **Rationale:** Manual import is sufficient for initial adoption. Bulk import tooling is valuable for large-scale migrations but can be added later without blocking Phase 1 delivery.

12. **Q: Custom validation for complex rules?**
    - **DECIDED:** Comprehensive validation in Phase 1
    - **Implementation:**
      - Project key validator (uppercase alphanumeric, underscore pattern)
      - Repository slug validator (lowercase alphanumeric, hyphen pattern)
      - Branch pattern validator (Git ref pattern)
      - SSH key format validator (validate key type and structure)
      - Permission level validator (enum validation)
    - **Rationale:** Custom validators provide immediate feedback before API calls, resulting in better error messages and faster feedback loops. The validation logic is straightforward and significantly improves user experience.

## Open Questions

None remaining - all architectural and implementation questions have been resolved.
