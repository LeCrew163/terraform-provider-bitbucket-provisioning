# Implementation Tasks

## 1. Project Setup & Infrastructure

- [ ] 1.1 Initialize uv project with pyproject.toml
- [ ] 1.2 Create src/ layout directory structure
- [ ] 1.3 Set up Python package structure (bitbucket_provisioning module)
- [ ] 1.4 Configure dependencies with uv (Click, Pydantic, httpx, PyYAML)
- [ ] 1.5 Create Makefile with common tasks (test, lint, generate-client)
- [ ] 1.6 Set up .gitignore for Python project
- [ ] 1.7 Configure Python 3.9+ compatibility settings
- [ ] 1.8 Create README.md with project overview and installation instructions

## 2. OpenAPI Client Generation

- [ ] 2.1 Download Bitbucket Data Center OpenAPI spec (v10.0)
- [ ] 2.2 Store OpenAPI spec in version control (specs/bitbucket-openapi.json)
- [ ] 2.3 Install openapi-python-client generator
- [ ] 2.4 Create generation script/Makefile target for client generation
- [ ] 2.5 Generate initial API client from OpenAPI spec
- [ ] 2.6 Create api/generated/ directory for generated code
- [ ] 2.7 Add generated client to .gitignore or commit based on decision
- [ ] 2.8 Create api/client.py wrapper for authentication and base URL management
- [ ] 2.9 Implement retry logic in API wrapper
- [ ] 2.10 Implement standardized error handling in API wrapper
- [ ] 2.11 Create API client factory function
- [ ] 2.12 Add API version compatibility checking

## 3. Configuration Management

- [ ] 3.1 Create config/models.py with Pydantic models for all capabilities
- [ ] 3.2 Implement ProjectConfig model with all capability fields
- [ ] 3.3 Implement PermissionsConfig model (users, groups)
- [ ] 3.4 Implement BranchPermissionRule model
- [ ] 3.5 Implement AccessKeyConfig model
- [ ] 3.6 Implement WorkflowConfig model (Git Flow, GitHub Flow)
- [ ] 3.7 Implement HooksConfig model
- [ ] 3.8 Implement DefaultReviewersConfig model
- [ ] 3.9 Create config/loader.py for YAML file loading
- [ ] 3.10 Implement modular YAML file reference resolution (relative paths)
- [ ] 3.11 Implement modular YAML file reference resolution (absolute paths)
- [ ] 3.12 Implement circular reference detection for file references
- [ ] 3.13 Create config/validator.py for configuration validation
- [ ] 3.14 Implement JSON Schema generation from Pydantic models
- [ ] 3.15 Create configuration file discovery logic (precedence rules)
- [ ] 3.16 Implement environment variable loading (BITBUCKET_URL, BITBUCKET_TOKEN)
- [ ] 3.17 Add configuration validation error messages with clear explanations

## 4. CLI Framework

- [ ] 4.1 Create cli/__init__.py with Click application setup
- [ ] 4.2 Implement main CLI entrypoint
- [ ] 4.3 Create project command group
- [ ] 4.4 Implement 'project apply' command
- [ ] 4.5 Implement 'project validate' command
- [ ] 4.6 Add --dry-run flag support to apply command
- [ ] 4.7 Add --json flag for CI/CD mode output
- [ ] 4.8 Add --verbose/-v flag for debug logging
- [ ] 4.9 Add --quiet/-q flag for minimal output
- [ ] 4.10 Implement configuration file path argument/option
- [ ] 4.11 Create output formatters (human-readable vs JSON)
- [ ] 4.12 Implement exit code handling (0=success, 1=error, 2=usage)
- [ ] 4.13 Add version command showing tool and API versions

## 5. Core Infrastructure

- [ ] 5.1 Create utils/logging.py with structured logging setup
- [ ] 5.2 Configure logging levels (INFO, DEBUG, WARNING)
- [ ] 5.3 Create utils/exceptions.py with exception hierarchy
- [ ] 5.4 Implement BitbucketProvisioningError base exception
- [ ] 5.5 Implement ConfigurationError exception
- [ ] 5.6 Implement ValidationError exception
- [ ] 5.7 Implement APIError and subclasses (AuthenticationError, ResourceNotFoundError)
- [ ] 5.8 Create connectivity pre-flight check function
- [ ] 5.9 Implement Bitbucket API version detection
- [ ] 5.10 Create diff utility for showing configuration changes

## 6. Base Provisioner Infrastructure

- [ ] 6.1 Create provisioners/__init__.py
- [ ] 6.2 Create provisioners/base.py with BaseProvisioner abstract class
- [ ] 6.3 Implement apply() abstract method signature
- [ ] 6.4 Implement validate() abstract method signature
- [ ] 6.5 Implement show() abstract method signature
- [ ] 6.6 Create ProvisionerResult model for standardized return values
- [ ] 6.7 Implement idempotency checking utilities
- [ ] 6.8 Create change detection and diff generation utilities

## 7. Project Creation Provisioner

- [ ] 7.1 Create provisioners/project_provisioner.py
- [ ] 7.2 Implement project creation via Bitbucket API
- [ ] 7.3 Implement project existence checking
- [ ] 7.4 Implement project update logic (description, visibility)
- [ ] 7.5 Implement project visibility configuration (public/private)
- [ ] 7.6 Implement project avatar upload (URL and file path)
- [ ] 7.7 Add validation for project key format
- [ ] 7.8 Implement idempotent project creation/update
- [ ] 7.9 Add dry-run mode support for project operations
- [ ] 7.10 Implement detailed operation result reporting

## 8. Project Permissions Provisioner

- [ ] 8.1 Create provisioners/permissions_provisioner.py
- [ ] 8.2 Implement user permission listing from Bitbucket
- [ ] 8.3 Implement group permission listing from Bitbucket
- [ ] 8.4 Implement user existence validation
- [ ] 8.5 Implement group existence validation
- [ ] 8.6 Implement user permission granting (PROJECT_READ, PROJECT_WRITE, PROJECT_ADMIN)
- [ ] 8.7 Implement group permission granting
- [ ] 8.8 Implement user permission revocation
- [ ] 8.9 Implement group permission revocation
- [ ] 8.10 Implement permission change detection and diff
- [ ] 8.11 Add dry-run mode support for permission operations
- [ ] 8.12 Implement idempotent permission management

## 9. Branch Permissions Provisioner

- [ ] 9.1 Create provisioners/branch_permissions_provisioner.py
- [ ] 9.2 Implement branch permission rule listing from Bitbucket
- [ ] 9.3 Implement branch pattern matching logic
- [ ] 9.4 Implement rule creation for "no-deletes" restriction type
- [ ] 9.5 Implement rule creation for "fast-forward-only" restriction type
- [ ] 9.6 Implement rule creation for "read-only" restriction type
- [ ] 9.7 Implement rule creation for "pull-request-only" restriction type
- [ ] 9.8 Implement exempted users/groups configuration
- [ ] 9.9 Implement rule deletion
- [ ] 9.10 Implement rule update logic
- [ ] 9.11 Add rule precedence and ordering logic
- [ ] 9.12 Implement change detection and diff for branch rules
- [ ] 9.13 Add dry-run mode support for branch permission operations
- [ ] 9.14 Implement idempotent rule management

## 10. Access Keys Provisioner

- [ ] 10.1 Create provisioners/access_keys_provisioner.py
- [ ] 10.2 Implement SSH key listing from Bitbucket project
- [ ] 10.3 Implement SSH key format validation
- [ ] 10.4 Implement SSH key type validation (RSA, ED25519, etc.)
- [ ] 10.5 Implement SSH key addition with label
- [ ] 10.6 Implement SSH key deletion
- [ ] 10.7 Implement SSH key label update
- [ ] 10.8 Support inline SSH keys in YAML
- [ ] 10.9 Support SSH key loading from file paths
- [ ] 10.10 Implement key fingerprint calculation and display
- [ ] 10.11 Implement duplicate key detection
- [ ] 10.12 Add permission configuration (read-only vs read-write)
- [ ] 10.13 Implement change detection and diff for access keys
- [ ] 10.14 Add dry-run mode support for access key operations
- [ ] 10.15 Implement idempotent key management

## 11. Branch Workflow Provisioner

- [ ] 11.1 Create provisioners/workflow_provisioner.py
- [ ] 11.2 Implement branching model listing from Bitbucket
- [ ] 11.3 Implement Git Flow configuration
- [ ] 11.4 Implement GitHub Flow configuration
- [ ] 11.5 Implement main/production branch name configuration
- [ ] 11.6 Implement development branch name configuration (Git Flow)
- [ ] 11.7 Implement branch prefix configuration (feature, release, hotfix)
- [ ] 11.8 Implement branching model enable/disable
- [ ] 11.9 Validate configured branches exist or can be created
- [ ] 11.10 Implement workflow template application (standard Git Flow, GitHub Flow)
- [ ] 11.11 Implement change detection and diff for workflow config
- [ ] 11.12 Add dry-run mode support for workflow operations
- [ ] 11.13 Implement idempotent workflow management

## 12. Hooks Configuration Provisioner

- [ ] 12.1 Create provisioners/hooks_provisioner.py
- [ ] 12.2 Implement project hooks listing from Bitbucket
- [ ] 12.3 Implement hook availability checking
- [ ] 12.4 Implement hook enable operation
- [ ] 12.5 Implement hook disable operation
- [ ] 12.6 Implement hook settings configuration (parameters)
- [ ] 12.7 Support built-in Bitbucket hooks configuration
- [ ] 12.8 Support third-party hook configuration by plugin key
- [ ] 12.9 Validate hook plugin existence and enabled status
- [ ] 12.10 Implement webhook creation with URL and events
- [ ] 12.11 Implement webhook authentication configuration
- [ ] 12.12 Implement webhook update and deletion
- [ ] 12.13 Implement hook settings validation against schema
- [ ] 12.14 Implement change detection and diff for hooks
- [ ] 12.15 Add dry-run mode support for hook operations
- [ ] 12.16 Implement idempotent hook management

## 13. Default Reviewers Provisioner

- [ ] 13.1 Create provisioners/reviewers_provisioner.py
- [ ] 13.2 Implement default reviewer rules listing from Bitbucket
- [ ] 13.3 Implement source branch pattern matching
- [ ] 13.4 Implement target branch pattern matching
- [ ] 13.5 Implement reviewer rule creation with users and groups
- [ ] 13.6 Implement required approval count configuration
- [ ] 13.7 Implement reviewer existence validation (users and groups)
- [ ] 13.8 Implement reviewer exclusion logic (exclude PR author)
- [ ] 13.9 Implement rule deletion
- [ ] 13.10 Implement rule update logic
- [ ] 13.11 Handle multiple matching rules and reviewer deduplication
- [ ] 13.12 Implement required vs optional reviewers configuration
- [ ] 13.13 Implement reviewer notification settings
- [ ] 13.14 Implement change detection and diff for reviewer rules
- [ ] 13.15 Add dry-run mode support for reviewer operations
- [ ] 13.16 Implement idempotent reviewer rule management

## 14. Orchestration & Integration

- [ ] 14.1 Create orchestrator to coordinate all provisioners
- [ ] 14.2 Implement provisioner execution order based on dependencies
- [ ] 14.3 Implement partial failure handling and reporting
- [ ] 14.4 Create consolidated diff output across all capabilities
- [ ] 14.5 Implement progress reporting during apply operations
- [ ] 14.6 Add rollback capability documentation (manual for Phase 1)
- [ ] 14.7 Create operation result aggregation and summary

## 15. Project Export (Phase 1.5)

- [ ] 21.1 Add export command to CLI command group
- [ ] 21.2 Implement project listing functionality
- [ ] 21.3 Implement project discovery with filters
- [ ] 21.4 Add show() method to BaseProvisioner interface
- [ ] 21.5 Implement show() in project creation provisioner
- [ ] 21.6 Implement show() in permissions provisioner
- [ ] 21.7 Implement show() in branch permissions provisioner
- [ ] 21.8 Implement show() in access keys provisioner
- [ ] 21.9 Implement show() in workflow provisioner
- [ ] 21.10 Implement show() in hooks provisioner
- [ ] 21.11 Implement show() in default reviewers provisioner
- [ ] 21.12 Create export orchestrator to aggregate all capabilities
- [ ] 21.13 Implement YAML serialization with comments and formatting
- [ ] 21.14 Add export metadata (timestamp, tool version, Bitbucket version)
- [ ] 21.15 Implement sensitive data handling (masking secrets)
- [ ] 21.16 Add verbose mode for export showing progress
- [ ] 21.17 Implement JSON export format option
- [ ] 21.18 Add file overwrite confirmation prompt
- [ ] 21.19 Implement export preview mode (dry-run for export)
- [ ] 21.20 Add batch export for multiple projects
- [ ] 21.21 Write unit tests for export functionality
- [ ] 21.22 Write integration tests for export command
- [ ] 21.23 Document export command usage and examples
- [ ] 21.24 Create migration guide using export

## 16. Testing - Unit Tests

- [ ] 21.1 Set up pytest configuration
- [ ] 21.2 Install pytest, pytest-mock, pytest-cov
- [ ] 21.3 Create tests/ directory structure mirroring src/
- [ ] 21.4 Write unit tests for configuration models
- [ ] 21.5 Write unit tests for YAML loading and validation
- [ ] 21.6 Write unit tests for modular YAML file references
- [ ] 21.7 Write unit tests for circular reference detection
- [ ] 21.8 Write unit tests for each provisioner (mocked API)
- [ ] 21.9 Write unit tests for API wrapper
- [ ] 21.10 Write unit tests for exception handling
- [ ] 21.11 Write unit tests for CLI commands using CliRunner
- [ ] 21.12 Achieve >80% code coverage

## 17. Testing - Integration Tests

- [ ] 21.1 Set up pytest-vcr for recording API interactions
- [ ] 21.2 Create integration test fixtures for test Bitbucket instance
- [ ] 21.3 Write integration tests for project creation
- [ ] 21.4 Write integration tests for permissions management
- [ ] 21.5 Write integration tests for branch permissions
- [ ] 21.6 Write integration tests for access keys
- [ ] 21.7 Write integration tests for workflow configuration
- [ ] 21.8 Write integration tests for hooks configuration
- [ ] 21.9 Write integration tests for default reviewers
- [ ] 21.10 Write end-to-end CLI integration tests

## 18. Documentation

- [ ] 21.1 Write comprehensive README.md with features and installation
- [ ] 21.2 Create docs/ directory for detailed documentation
- [ ] 21.3 Document YAML configuration schema with examples
- [ ] 21.4 Create example configurations for each capability
- [ ] 21.5 Create example for modular YAML file references
- [ ] 21.6 Document shared configuration patterns and best practices
- [ ] 21.7 Document CLI commands and options
- [ ] 21.8 Create troubleshooting guide
- [ ] 21.9 Document authentication setup (PAT vs HTTP Basic)
- [ ] 21.10 Document environment variable configuration
- [ ] 21.11 Create security best practices documentation
- [ ] 21.12 Document minimum Bitbucket Data Center version requirements
- [ ] 21.13 Create contribution guidelines
- [ ] 21.14 Generate API documentation from code

## 19. CI/CD Integration

- [ ] 21.1 Create example CI/CD pipeline configuration (GitHub Actions)
- [ ] 21.2 Create example CI/CD pipeline configuration (GitLab CI)
- [ ] 21.3 Document CI/CD integration patterns
- [ ] 21.4 Create example for secret management in CI/CD
- [ ] 21.5 Document --json flag usage in CI/CD
- [ ] 21.6 Create example for dry-run in CI/CD (PR validation)

## 20. Packaging & Distribution

- [ ] 21.1 Configure build settings in pyproject.toml
- [ ] 21.2 Set up CLI entrypoint in pyproject.toml
- [ ] 21.3 Create package metadata (description, authors, license)
- [ ] 21.4 Build wheel distribution with uv build
- [ ] 21.5 Test installation from built wheel
- [ ] 21.6 Document installation methods (PyPI, git, wheel)
- [ ] 21.7 Create release process documentation
- [ ] 21.8 Set up versioning strategy (semantic versioning)

## 21. Final Validation

- [ ] 21.1 Run full test suite and verify all tests pass
- [ ] 21.2 Test against real Bitbucket Data Center instance
- [ ] 21.3 Validate all CLI commands work as documented
- [ ] 21.4 Test dry-run mode for all operations
- [ ] 21.5 Test idempotent operations (apply twice, same result)
- [ ] 21.6 Test modular YAML configurations with file references
- [ ] 21.7 Verify error messages are clear and actionable
- [ ] 21.8 Test CI/CD integration with --json output
- [ ] 21.9 Review and validate all documentation
- [ ] 21.10 Create demo video or tutorial walkthrough
