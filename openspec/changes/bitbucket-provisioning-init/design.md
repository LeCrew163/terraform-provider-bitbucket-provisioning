## Context

This is a greenfield Python project to automate Bitbucket Data Center project provisioning. Currently, DevOps/Platform Engineers manually configure projects through the Bitbucket UI, leading to inconsistency and errors. This tool will use the official Bitbucket Data Center REST API v3 (OpenAPI spec) to provision projects declaratively from YAML configuration files.

**Constraints:**
- Must work with Bitbucket Data Center (not Cloud)
- Must support CI/CD integration
- Must be maintainable by DevOps team familiar with Python
- API client must stay synchronized with Bitbucket API changes

**Stakeholders:**
- Primary users: DevOps/Platform Engineers
- Secondary users: CI/CD automation systems

## Goals / Non-Goals

**Goals:**
- Create production-ready Python CLI tool for Bitbucket project provisioning
- Define optimal project structure for maintainability and extensibility
- Establish OpenAPI client generation workflow
- Design modular architecture supporting future repository management
- Enable both interactive CLI usage and CI/CD automation

**Non-Goals:**
- Repository management (Phase 2 - future work)
- Bitbucket Cloud support
- Web UI or REST API service (CLI only)
- Migration from existing Bitbucket projects
- Backup/restore functionality

## Decisions

### 1. Project Structure: Src Layout

**Decision:** Use `src/` layout for Python package structure.

**Rationale:**
- Prevents accidental imports of uninstalled package during development
- Clear separation between source code and tests/docs
- Industry standard for modern Python projects
- Better support for editable installs

**Structure:**
```
bitbucket-provisioning/
├── src/
│   └── bitbucket_provisioning/
│       ├── __init__.py
│       ├── cli/              # CLI commands and entrypoint
│       ├── config/           # YAML parsing and validation
│       ├── api/              # Generated OpenAPI client wrapper
│       ├── provisioners/     # Business logic for each capability
│       └── utils/            # Shared utilities
├── tests/
├── docs/
├── openspec/                 # OpenSpec artifacts
├── pyproject.toml
└── README.md
```

**Alternatives Considered:**
- Flat layout: Rejected due to import issues and less clear structure
- Separate packages per capability: Rejected as over-engineered for initial scope

### 2. OpenAPI Client Generation: openapi-python-client

**Decision:** Use `openapi-python-client` generator with custom wrapper layer.

**Rationale:**
- Generates modern, type-annotated Python code (uses `attrs`, `httpx`)
- Better developer experience than alternatives (mypy support, IDE autocomplete)
- Actively maintained
- Produces cleaner code than Java-based generators (openapi-generator)

**Implementation:**
- Generated client lives in `src/bitbucket_provisioning/api/generated/`
- Thin wrapper layer in `src/bitbucket_provisioning/api/client.py` for:
  - Authentication configuration
  - Error handling standardization
  - Retry logic
  - Base URL management

**Regeneration workflow:**
```bash
# Makefile target
make generate-client
```

**Alternatives Considered:**
- `openapi-generator`: Rejected - generates overly verbose code, Java-style patterns
- Manual API client: Rejected - too much maintenance burden, error-prone
- `datamodel-code-generator` + manual client: Rejected - more work, less benefit

### 3. CLI Framework: Click

**Decision:** Use Click for CLI implementation.

**Rationale:**
- Most popular Python CLI framework, mature and stable
- Excellent support for nested command groups
- Built-in help generation and parameter validation
- Easy testing with `CliRunner`
- Rich plugin ecosystem

**Command Structure:**
```bash
bitbucket-provision project create <yaml-file>
bitbucket-provision project apply <yaml-file>
bitbucket-provision project validate <yaml-file>
bitbucket-provision project show <project-key>
```

**Alternatives Considered:**
- `argparse`: Rejected - too verbose, harder to organize nested commands
- `typer`: Rejected - less mature, smaller ecosystem, Click integration is proven
- `fire`: Rejected - less explicit, harder to maintain

### 4. Configuration Schema: Pydantic

**Decision:** Use Pydantic v2 for YAML schema validation and configuration models.

**Rationale:**
- Industry standard for data validation in Python
- Type-safe with excellent IDE support
- Rich validation features (custom validators, field constraints)
- JSON Schema generation for documentation
- Great error messages out of the box

**Configuration Model Structure:**
```python
# src/bitbucket_provisioning/config/models.py
class ProjectConfig(BaseModel):
    project: ProjectSettings
    permissions: Optional[PermissionsConfig]
    branch_permissions: Optional[List[BranchPermissionRule]]
    access_keys: Optional[List[AccessKeyConfig]]
    # ... other capabilities
```

**Alternatives Considered:**
- `marshmallow`: Rejected - less type-safe, more verbose
- `cerberus`: Rejected - dictionary-based, less Pythonic
- Manual validation: Rejected - error-prone, poor developer experience

### 5. Dependency Management: Poetry

**Decision:** Use Poetry for dependency management and packaging.

**Rationale:**
- Modern Python packaging (PEP 517/518 compliant)
- Lock file ensures reproducible builds
- Simplified dependency resolution
- Built-in virtual environment management
- Easy publish to PyPI or private package repository

**Alternatives Considered:**
- `pip` + `requirements.txt`: Rejected - no lock file, manual version management
- `pipenv`: Rejected - slower, less active development
- `uv`: Considered but too new, less mature ecosystem

### 6. Modular YAML Configuration: File References

**Decision:** Support modular YAML configurations with path-based file references.

**Rationale:**
- Better organization for complex configurations
- Reusability of common configs across multiple projects
- Easier code review and version control (separate files for different concerns)
- Team ownership (different teams can maintain different config files)
- Flexibility: small configs inline, complex configs external

**Structure:**
```yaml
# project.yaml (main config)
project:
  key: MYPROJ
  name: My Project
  description: "Example project"
  visibility: private

# Inline for simple configs
permissions:
  users:
    - username: john.doe
      permission: PROJECT_ADMIN

# External references for complex configs
branch_permissions: ./configs/branch-permissions.yaml
hooks: /opt/shared/configs/standard-hooks.yaml
default_reviewers: ../shared/reviewers/team-leads.yaml
```

**File Reference Resolution:**
- Relative paths: Resolved relative to the main config file location
- Absolute paths: Used as-is from filesystem root
- Both supported for maximum flexibility

**Use Cases:**
- **Per-capability separation**: `permissions.yaml`, `hooks.yaml`, etc.
- **Shared configs**: Common configurations shared across projects
  - Standard permission templates
  - Common hook configurations
  - Organization-wide reviewer rules
- **Organization structure**: Mirror team/project organization in config directory layout

**Validation:**
- Validate all referenced files exist before parsing
- Parse and validate each referenced file independently
- Merge referenced configs into main configuration model
- Circular reference detection

**Alternatives Considered:**
- YAML tags/directives (`!include`): Rejected - less standard, requires custom YAML parser
- Import statements: Rejected - more complex, less intuitive
- Always monolithic: Rejected - poor for large/complex projects
- Always modular: Rejected - overkill for simple projects

### 7. Single Bitbucket Instance: One Instance Per Tool Invocation

**Decision:** Support only one Bitbucket instance per configuration/invocation.

**Rationale:**
- Simplifies configuration structure
- Clearer authentication setup
- Most common use case (single company Bitbucket instance)
- Reduces complexity in error handling and logging

**Configuration:**
```bash
# Single instance configured via environment
export BITBUCKET_URL="https://bitbucket.example.com"
```

**Future Consideration:**
- Multi-instance can be achieved by multiple config files
- Or by running tool multiple times with different env vars

**Alternatives Considered:**
- Multiple instances in one config: Rejected - added complexity for uncommon use case
- Instance per YAML file: Considered - could be Phase 2 enhancement

### 8. Authentication Strategy: Token-Based

**Decision:** Use Bitbucket Personal Access Tokens or HTTP Basic Auth.

**Rationale:**
- Data Center supports both PAT and HTTP Basic Auth
- PATs are more secure (scoped, revocable)
- No OAuth flow needed (server-to-server)

**Configuration:**
```bash
# Environment variables
export BITBUCKET_URL="https://bitbucket.example.com"
export BITBUCKET_TOKEN="<personal-access-token>"
# or
export BITBUCKET_USERNAME="<username>"
export BITBUCKET_PASSWORD="<password>"
```

**Alternatives Considered:**
- OAuth: Rejected - unnecessary complexity for service-to-service
- SSH keys: Rejected - not supported by REST API

### 9. Module Organization: Capability-Based Provisioners

**Decision:** Organize business logic into capability-specific provisioner modules.

**Rationale:**
- Clear separation of concerns
- Each capability maps to proposal capabilities
- Easy to test in isolation
- Supports phased implementation

**Structure:**
```python
# src/bitbucket_provisioning/provisioners/
├── __init__.py
├── base.py                    # Base provisioner class
├── project_provisioner.py     # project-creation capability
├── permissions_provisioner.py # project-permissions capability
├── branch_permissions_provisioner.py
├── access_keys_provisioner.py
├── workflow_provisioner.py
├── hooks_provisioner.py
└── reviewers_provisioner.py
```

**Each provisioner:**
- Inherits from `BaseProvisioner`
- Implements `apply()`, `validate()`, `show()` methods
- Uses generated API client
- Returns standardized result objects

**Alternatives Considered:**
- Single monolithic module: Rejected - hard to maintain, test
- Per-command organization: Rejected - duplicates logic across commands

### 10. Testing Strategy: Pytest with Multiple Layers

**Decision:** Multi-layer testing approach.

**Test Layers:**
1. **Unit tests**: Individual functions, provisioners (mocked API)
2. **Integration tests**: Real API calls against test Bitbucket instance
3. **CLI tests**: End-to-end CLI commands using Click's `CliRunner`
4. **Contract tests**: Validate API client against OpenAPI spec

**Framework:** pytest with:
- `pytest-mock` for mocking
- `pytest-vcr` for recording/replaying HTTP interactions
- `pytest-cov` for coverage reporting

**Alternatives Considered:**
- unittest only: Rejected - more verbose, less powerful
- No integration tests: Rejected - risky without real API validation

### 11. Error Handling: Structured Exceptions

**Decision:** Define custom exception hierarchy with structured error messages.

**Exception Hierarchy:**
```python
BitbucketProvisioningError (base)
├── ConfigurationError
├── ValidationError
├── APIError
│   ├── AuthenticationError
│   ├── ResourceNotFoundError
│   └── RateLimitError
└── ProvisioningError
```

**Error Output:**
- Structured JSON in CI/CD mode (`--json`)
- Human-readable messages in interactive mode
- Exit codes aligned with conventions (0=success, 1=error, 2=usage error)

**Alternatives Considered:**
- Generic exceptions: Rejected - hard to handle specific cases
- Error codes only: Rejected - less Pythonic

### 12. Logging: Structured Logging with Levels

**Decision:** Use Python's `logging` module with structured output.

**Configuration:**
- Default: INFO level to stderr
- Verbose mode (`-v`): DEBUG level
- Quiet mode (`-q`): WARNING level only
- CI/CD mode: Structured JSON logs

**Alternatives Considered:**
- `loguru`: Rejected - overkill for CLI tool
- Print statements: Rejected - not configurable, no levels

### 13. Configuration File Discovery

**Decision:** Support multiple configuration sources with precedence.

**Precedence (highest to lowest):**
1. CLI arguments (`--file config.yaml`)
2. Environment variable (`BITBUCKET_CONFIG`)
3. Current directory (`./bitbucket.yaml` or `./bitbucket.yml`)
4. User config directory (`~/.config/bitbucket-provisioning/config.yaml`)

**Rationale:**
- Flexible for different use cases (local dev, CI/CD, shared configs)
- Follows XDG Base Directory specification
- Clear precedence rules prevent confusion

**Alternatives Considered:**
- Config file required: Rejected - less flexible
- Only CLI arguments: Rejected - poor CI/CD experience

### 14. Idempotent Operations: Always Update

**Decision:** `apply` command is fully idempotent - creates or updates projects based on YAML state.

**Rationale:**
- Best CI/CD experience - same command works every time
- Declarative approach - YAML is single source of truth
- Follows Infrastructure-as-Code best practices (Terraform, Kubernetes)
- GitOps friendly - commit config changes, run same command

**Behavior:**
```bash
# First run: creates project
$ bitbucket-provision project apply config.yaml
✓ Created project 'MYPROJ'

# Second run: updates if needed, no-op if already correct
$ bitbucket-provision project apply config.yaml
✓ Project 'MYPROJ' already up to date

# After config change: updates only what changed
$ bitbucket-provision project apply config.yaml
✓ Updated project 'MYPROJ'
  ✓ Updated description
  ✓ Added 2 branch permissions
```

**Safeguards:**
- `--dry-run` flag: Preview changes without applying
- Detailed diff output: Show exactly what will change
- `validate` command: Check config before applying

**Implementation:**
- Each provisioner implements idempotent logic
- Check current state before making changes
- Only call API for actual changes (avoid unnecessary updates)
- Return structured diff of changes made

**Alternatives Considered:**
- Error on conflict: Rejected - poor CI/CD experience, not declarative
- Hybrid with `--update` flag: Rejected - adds complexity, conditional logic in CI
- Separate create/update commands: Rejected - forces users to track state externally

## Risks / Trade-offs

### Risk 1: OpenAPI Spec Changes
**Risk:** Bitbucket updates API, breaking generated client.

**Mitigation:**
- Pin OpenAPI spec version in repository
- Include spec file in version control
- Document regeneration process
- Add CI checks for spec changes
- Version lock generated client with semantic versioning

### Risk 2: Bitbucket Data Center Version Compatibility
**Risk:** Different Bitbucket versions have different API capabilities.

**Mitigation:**
- Document minimum supported Bitbucket version (10.0+)
- Add version detection in client
- Graceful degradation for missing features
- Clear error messages for unsupported versions

### Risk 3: Authentication Token Management
**Risk:** Tokens in environment variables could leak in CI logs.

**Mitigation:**
- Document best practices (use CI secrets)
- Support external credential providers (future)
- Never log token values
- Warn if token detected in command output

### Risk 4: Partial Application Failures
**Risk:** Project created but permissions fail - inconsistent state.

**Mitigation:**
- Implement idempotent operations where possible
- Dry-run mode (`--dry-run`) to preview changes
- Clear error messages indicating which step failed
- Consider rollback functionality (future enhancement)

### Risk 5: YAML Configuration Complexity
**Risk:** YAML becomes too complex, hard to write/maintain.

**Mitigation:**
- Provide comprehensive examples
- JSON Schema generation for editor support
- `validate` command for pre-flight checks
- Template generation command (future)

### Trade-off 1: Generated Code in Repository
**Decision:** Commit generated API client to version control.

**Trade-off:**
- Pro: Reproducible builds, no generation step required by users
- Con: Larger repository, potential merge conflicts

**Rationale:** Reproducibility and ease of installation outweigh repository size concerns.

### Trade-off 2: Python 3.9+ Requirement
**Decision:** Target Python 3.9 as minimum version.

**Trade-off:**
- Pro: Modern type hints, better performance, active support
- Con: Excludes older Python users

**Rationale:** Python 3.9 released Sept 2020, good balance of features vs. compatibility.

## Migration Plan

**N/A** - This is a greenfield project with no existing codebase to migrate.

**Deployment:**
1. Package as Python wheel
2. Publish to internal PyPI repository (or install via git)
3. Document installation: `pip install bitbucket-provisioning`
4. Provide example configurations
5. Team onboarding sessions

**Rollback:**
- N/A for tool deployment (users pin versions)
- Changes made by tool to Bitbucket: Manual rollback initially, automated rollback in future version

## Resolved Questions

1. **Q: Should we support multiple Bitbucket instances in one config?**
   - **DECIDED:** No - single instance only for Phase 1
   - Simplifies configuration and authentication
   - Can be added in future if needed

2. **Q: How to handle existing projects - update vs. error?**
   - **DECIDED:** Always idempotent (see Decision #14)
   - `apply` command creates or updates based on YAML state
   - Includes `--dry-run` for safety

3. **Q: How to handle sensitive data in YAML (e.g., webhook secrets)?**
   - **DECIDED:** Phase 1 - no secret management (plain text with security warnings)
   - **DECIDED:** Phase 2 - AWS SSM Parameter Store or Secrets Manager integration
   - Config structure designed to support future secret references: `${aws:ssm:/path}`

4. **Q: Should generated API client be published as separate package?**
   - **DECIDED:** Keep internal for now, revisit if demand arises
   - Reduces maintenance overhead
   - Can extract later if needed

5. **Q: Should we validate Bitbucket connectivity before applying config?**
   - **DECIDED:** Yes - implement in Phase 1 (base implementation)
   - Pre-flight check before any operations
   - Test authentication and API availability
   - Fail fast with clear error messages
   - Validates: URL reachable, credentials valid, API version compatible

6. **Q: Scope of `project show` command - how much detail?**
   - **DECIDED:** Future enhancement - not a priority for Phase 1
   - Could be added later with options like: `show --format=yaml` to export current state
   - Consider summary vs detailed views with flags

7. **Q: Should we support partial config application?**
   - **DECIDED:** Future enhancement - not a priority for Phase 1
   - Could be added later: `apply --only=permissions` or separate YAML files
   - Base implementation applies entire config

8. **Q: How to handle drift detection?**
   - **DECIDED:** High priority for Phase 2 (after base implementation)
   - Implement `diff` command to compare YAML vs actual Bitbucket state
   - Show what manual changes were made outside of the tool
   - Essential for maintaining configuration as source of truth

## Open Questions

None remaining - all key architectural decisions have been made.
