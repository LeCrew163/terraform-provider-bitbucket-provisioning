# Implementation Tasks

## 1. Project Setup & Infrastructure

- [ ] 1.1 Initialize Go module with go.mod
- [ ] 1.2 Create standard Terraform provider directory structure
- [ ] 1.3 Set up internal/ package organization
- [ ] 1.4 Configure Go dependencies (Terraform Plugin Framework, testing libraries)
- [ ] 1.5 Create Makefile with common tasks (build, test, generate, docs)
- [ ] 1.6 Set up .gitignore for Go project
- [ ] 1.7 Configure Go version requirement (1.21+)
- [ ] 1.8 Create main.go provider entrypoint
- [ ] 1.9 Create README.md with project overview and development setup
- [ ] 1.10 Set up .goreleaser.yml for releases

## 2. OpenAPI Client Generation

- [ ] 2.1 Download Bitbucket Data Center OpenAPI spec (v10.0)
- [ ] 2.2 Store OpenAPI spec in version control (specs/bitbucket-openapi.json)
- [ ] 2.3 Install openapi-generator CLI
- [ ] 2.4 Create generation script/Makefile target for Go client generation
- [ ] 2.5 Generate initial API client from OpenAPI spec
- [ ] 2.6 Create internal/client/generated/ directory for generated code
- [ ] 2.7 Commit generated client to version control
- [ ] 2.8 Create internal/client/client.go wrapper for configuration
- [ ] 2.9 Implement authentication support (PAT and HTTP Basic)
- [ ] 2.10 Implement retry logic with exponential backoff
- [ ] 2.11 Implement error handling and translation to Terraform diagnostics
- [ ] 2.12 Create client factory function with configuration
- [ ] 2.13 Add API version compatibility checking
- [ ] 2.14 Implement context propagation for all client operations

## 3. Provider Core Implementation

- [ ] 3.1 Create internal/provider/provider.go
- [ ] 3.2 Implement provider schema with configuration attributes
- [ ] 3.3 Implement Configure method for provider initialization
- [ ] 3.4 Add base_url configuration attribute
- [ ] 3.5 Add token configuration attribute (marked sensitive)
- [ ] 3.6 Add username/password configuration attributes (marked sensitive)
- [ ] 3.7 Add insecure_skip_verify configuration attribute
- [ ] 3.8 Add timeout configuration attribute
- [ ] 3.9 Implement environment variable support for configuration
- [ ] 3.10 Add validation for required configuration
- [ ] 3.11 Implement provider metadata (name, version)
- [ ] 3.12 Register resources in provider Resources method
- [ ] 3.13 Register data sources in provider DataSources method
- [ ] 3.14 Add provider-level diagnostics and error handling
- [ ] 3.15 Implement pre-flight connectivity check

## 4. Project Resource

- [ ] 4.1 Create internal/provider/resource_project.go
- [ ] 4.2 Define project resource schema with all attributes
- [ ] 4.3 Implement resource Metadata method
- [ ] 4.4 Implement Create method for project creation
- [ ] 4.5 Implement Read method for state refresh
- [ ] 4.6 Implement Update method for project updates
- [ ] 4.7 Implement Delete method for project deletion
- [ ] 4.8 Add validation for project key format
- [ ] 4.9 Implement computed ID attribute
- [ ] 4.10 Add RequiresReplace for immutable key attribute
- [ ] 4.11 Implement ImportState method with key-based import
- [ ] 4.12 Handle project visibility (public/private)
- [ ] 4.13 Handle project description (optional)
- [ ] 4.14 Add error handling for duplicate keys
- [ ] 4.15 Add error handling for permission errors
- [ ] 4.16 Implement avatar configuration (optional, future)
- [ ] 4.17 Write unit tests for project resource schema
- [ ] 4.18 Write acceptance tests for project lifecycle

## 5. Repository Resource

- [ ] 5.1 Create internal/provider/resource_repository.go
- [ ] 5.2 Define repository resource schema
- [ ] 5.3 Implement Create method for repository creation
- [ ] 5.4 Implement Read method for state refresh
- [ ] 5.5 Implement Update method for repository updates
- [ ] 5.6 Implement Delete method for repository deletion
- [ ] 5.7 Add project_key reference attribute
- [ ] 5.8 Add repository slug attribute (RequiresReplace)
- [ ] 5.9 Add repository name attribute
- [ ] 5.10 Add repository description attribute (optional)
- [ ] 5.11 Add fork configuration (is_forkable, forks_enabled)
- [ ] 5.12 Add default branch attribute
- [ ] 5.13 Implement ImportState with "project-key/repo-slug" pattern
- [ ] 5.14 Handle repository visibility (public/private)
- [ ] 5.15 Add validation for slug format
- [ ] 5.16 Write unit tests for repository resource
- [ ] 5.17 Write acceptance tests for repository lifecycle
- [ ] 5.18 Test repository updates and modifications

## 6. Project Permissions Resource

- [ ] 6.1 Create internal/provider/resource_project_permissions.go
- [ ] 6.2 Define project permissions schema with user/group blocks
- [ ] 6.3 Implement Create method (grant all permissions)
- [ ] 6.4 Implement Read method (list current permissions)
- [ ] 6.5 Implement Update method (reconcile permission changes)
- [ ] 6.6 Implement Delete method (revoke all permissions)
- [ ] 6.7 Add project_key reference attribute
- [ ] 6.8 Add user block with name and permission attributes
- [ ] 6.9 Add group block with name and permission attributes
- [ ] 6.10 Support PROJECT_READ, PROJECT_WRITE, PROJECT_ADMIN permission levels
- [ ] 6.11 Validate user and group existence
- [ ] 6.12 Implement ImportState for project permissions
- [ ] 6.13 Handle permission additions and removals
- [ ] 6.14 Add error handling for invalid users/groups
- [ ] 6.15 Write unit tests for permissions resource
- [ ] 6.16 Write acceptance tests for permission management
- [ ] 6.17 Test permission updates and reconciliation

## 7. Branch Permissions Resource

- [ ] 7.1 Create internal/provider/resource_branch_permissions.go
- [ ] 7.2 Define branch permissions schema with restriction blocks
- [ ] 7.3 Implement Create method for restriction creation
- [ ] 7.4 Implement Read method for listing restrictions
- [ ] 7.5 Implement Update method for restriction updates
- [ ] 7.6 Implement Delete method for restriction deletion
- [ ] 7.7 Add project_key reference attribute
- [ ] 7.8 Add restriction block schema
- [ ] 7.9 Support restriction types (fast-forward-only, no-deletes, read-only, pull-request-only)
- [ ] 7.10 Add branch_pattern attribute (supports wildcards)
- [ ] 7.11 Add exempted_users list attribute
- [ ] 7.12 Add exempted_groups list attribute
- [ ] 7.13 Implement ImportState for branch permissions
- [ ] 7.14 Handle multiple restrictions per project
- [ ] 7.15 Validate branch pattern syntax
- [ ] 7.16 Write unit tests for branch permissions resource
- [ ] 7.17 Write acceptance tests for branch permission rules
- [ ] 7.18 Test restriction updates and deletions

## 8. Project Access Keys Resource

- [ ] 8.1 Create internal/provider/resource_project_access_keys.go
- [ ] 8.2 Define access keys schema with key blocks
- [ ] 8.3 Implement Create method for adding SSH keys
- [ ] 8.4 Implement Read method for listing keys
- [ ] 8.5 Implement Update method for key updates
- [ ] 8.6 Implement Delete method for key removal
- [ ] 8.7 Add project_key reference attribute
- [ ] 8.8 Add key block with public_key and label attributes
- [ ] 8.9 Add permission attribute (read vs read-write)
- [ ] 8.10 Validate SSH key format
- [ ] 8.11 Support RSA, ED25519, and other key types
- [ ] 8.12 Mark public_key as sensitive attribute
- [ ] 8.13 Implement key fingerprint computation
- [ ] 8.14 Implement ImportState for access keys
- [ ] 8.15 Handle duplicate key detection
- [ ] 8.16 Write unit tests for access keys resource
- [ ] 8.17 Write acceptance tests for SSH key management
- [ ] 8.18 Test key addition, update, and removal

## 9. Project Data Source

- [ ] 9.1 Create internal/provider/data_source_project.go
- [ ] 9.2 Define project data source schema
- [ ] 9.3 Implement Read method to query project by key
- [ ] 9.4 Return project attributes (name, description, visibility, id)
- [ ] 9.5 Add validation for required key attribute
- [ ] 9.6 Handle not found errors gracefully
- [ ] 9.7 Write unit tests for project data source
- [ ] 9.8 Write acceptance tests for project queries
- [ ] 9.9 Test with existing and non-existent projects

## 10. Repository Data Source

- [ ] 10.1 Create internal/provider/data_source_repository.go
- [ ] 10.2 Define repository data source schema
- [ ] 10.3 Implement Read method to query repository by project and slug
- [ ] 10.4 Return repository attributes
- [ ] 10.5 Add project_key and slug filter attributes
- [ ] 10.6 Handle not found errors
- [ ] 10.7 Write unit tests for repository data source
- [ ] 10.8 Write acceptance tests for repository queries

## 11. User Data Source

- [ ] 11.1 Create internal/provider/data_source_user.go
- [ ] 11.2 Define user data source schema
- [ ] 11.3 Implement Read method to query user by username
- [ ] 11.4 Return user attributes (id, name, email, display_name)
- [ ] 11.5 Handle not found errors
- [ ] 11.6 Write unit tests for user data source
- [ ] 11.7 Write acceptance tests for user queries

## 12. Group Data Source

- [ ] 12.1 Create internal/provider/data_source_group.go
- [ ] 12.2 Define group data source schema
- [ ] 12.3 Implement Read method to query group by name
- [ ] 12.4 Return group attributes
- [ ] 12.5 Handle not found errors
- [ ] 12.6 Write unit tests for group data source
- [ ] 12.7 Write acceptance tests for group queries

## 13. Custom Validators

- [ ] 13.1 Create internal/validators/ package
- [ ] 13.2 Implement project key validator (uppercase alphanumeric, underscore)
- [ ] 13.3 Implement repository slug validator (lowercase alphanumeric, dash)
- [ ] 13.4 Implement SSH key format validator
- [ ] 13.5 Implement branch pattern validator
- [ ] 13.6 Implement permission level validator
- [ ] 13.7 Write unit tests for all validators

## 14. Testing - Unit Tests

- [ ] 14.1 Set up testing infrastructure
- [ ] 14.2 Install testing dependencies (terraform-plugin-testing)
- [ ] 14.3 Create test helpers and utilities
- [ ] 14.4 Write unit tests for provider configuration
- [ ] 14.5 Write unit tests for client wrapper
- [ ] 14.6 Write unit tests for error handling
- [ ] 14.7 Write unit tests for all validators
- [ ] 14.8 Write unit tests for import ID parsing
- [ ] 14.9 Write unit tests for schema definitions
- [ ] 14.10 Achieve >80% code coverage for non-acceptance tests

## 15. Testing - Integration Tests

- [ ] 15.1 Set up mock Bitbucket API server using httptest
- [ ] 15.2 Create test fixtures for API responses
- [ ] 15.3 Write integration tests for client operations
- [ ] 15.4 Test authentication flows (PAT and Basic)
- [ ] 15.5 Test retry logic and error handling
- [ ] 15.6 Test rate limiting behavior
- [ ] 15.7 Test API error translation to diagnostics

## 16. Testing - Acceptance Tests

- [ ] 16.1 Set up acceptance test framework
- [ ] 16.2 Create test Bitbucket Data Center instance (or use existing)
- [ ] 16.3 Implement testAccPreCheck function
- [ ] 16.4 Create provider factory for tests
- [ ] 16.5 Write acceptance tests for project resource (CRUD)
- [ ] 16.6 Write acceptance tests for repository resource (CRUD)
- [ ] 16.7 Write acceptance tests for project permissions
- [ ] 16.8 Write acceptance tests for branch permissions
- [ ] 16.9 Write acceptance tests for access keys
- [ ] 16.10 Write acceptance tests for all data sources
- [ ] 16.11 Test import functionality for all resources
- [ ] 16.12 Test update scenarios and state refresh
- [ ] 16.13 Test error conditions and edge cases
- [ ] 16.14 Document acceptance test setup and execution

## 17. Documentation

- [ ] 17.1 Create templates/ directory
- [ ] 17.2 Write provider documentation template (index.md.tmpl)
- [ ] 17.3 Write resource documentation templates
- [ ] 17.4 Write data source documentation templates
- [ ] 17.5 Create examples/ directory structure
- [ ] 17.6 Create provider configuration examples
- [ ] 17.7 Create project resource examples
- [ ] 17.8 Create repository resource examples
- [ ] 17.9 Create permissions resource examples
- [ ] 17.10 Create branch permissions examples
- [ ] 17.11 Create access keys examples
- [ ] 17.12 Create data source examples
- [ ] 17.13 Create complete module examples
- [ ] 17.14 Install tfplugindocs tool
- [ ] 17.15 Generate documentation with tfplugindocs
- [ ] 17.16 Review and validate generated documentation
- [ ] 17.17 Write migration guide for existing infrastructure
- [ ] 17.18 Write authentication setup guide
- [ ] 17.19 Write troubleshooting guide
- [ ] 17.20 Document supported Bitbucket versions

## 18. Advanced Features (Phase 2)

- [ ] 18.1 Create resource for project hooks configuration
- [ ] 18.2 Create resource for project webhooks
- [ ] 18.3 Create resource for default reviewers
- [ ] 18.4 Create resource for branch workflow configuration
- [ ] 18.5 Create resource for repository permissions
- [ ] 18.6 Create resource for repository hooks
- [ ] 18.7 Create data source for listing repositories
- [ ] 18.8 Create data source for listing projects
- [ ] 18.9 Write acceptance tests for advanced features
- [ ] 18.10 Document advanced resource usage

## 19. CI/CD Setup

- [ ] 19.1 Create GitHub Actions workflow for testing
- [ ] 19.2 Set up workflow for unit tests on PR
- [ ] 19.3 Set up workflow for linting (golangci-lint)
- [ ] 19.4 Set up workflow for acceptance tests (manual trigger)
- [ ] 19.5 Create release workflow using GoReleaser
- [ ] 19.6 Configure GPG signing for releases
- [ ] 19.7 Set up dependabot for dependency updates
- [ ] 19.8 Add status badges to README
- [ ] 19.9 Configure branch protection rules

## 20. Release Preparation

- [ ] 20.1 Create CHANGELOG.md
- [ ] 20.2 Create initial version (v0.1.0)
- [ ] 20.3 Set up GPG key for signing
- [ ] 20.4 Test GoReleaser locally
- [ ] 20.5 Create GitHub release
- [ ] 20.6 Verify release artifacts
- [ ] 20.7 Test provider installation from release
- [ ] 20.8 Update documentation with installation instructions

## 21. Terraform Registry Submission

- [ ] 21.1 Review Terraform Registry requirements
- [ ] 21.2 Ensure all documentation is complete
- [ ] 21.3 Verify examples are working
- [ ] 21.4 Create registry submission request
- [ ] 21.5 Respond to HashiCorp review feedback
- [ ] 21.6 Publish provider to Terraform Registry
- [ ] 21.7 Update README with registry installation instructions
- [ ] 21.8 Announce provider availability

## 22. Quality Assurance

- [ ] 22.1 Run full test suite and verify all tests pass
- [ ] 22.2 Test against real Bitbucket Data Center instance
- [ ] 22.3 Validate all resources work as documented
- [ ] 22.4 Test import for all resources
- [ ] 22.5 Test state refresh and drift detection
- [ ] 22.6 Verify error messages are clear and actionable
- [ ] 22.7 Test with Terraform Cloud
- [ ] 22.8 Test with Terraform Enterprise (if available)
- [ ] 22.9 Review and validate all documentation
- [ ] 22.10 Create demo repository with examples

## 23. Performance & Optimization

- [ ] 23.1 Profile provider performance
- [ ] 23.2 Optimize API client connection pooling
- [ ] 23.3 Implement request rate limiting
- [ ] 23.4 Add caching where appropriate
- [ ] 23.5 Test with large configurations (many resources)
- [ ] 23.6 Document performance best practices

## 24. Security Review

- [ ] 24.1 Review authentication implementation
- [ ] 24.2 Ensure sensitive values are properly marked
- [ ] 24.3 Verify no credentials in logs
- [ ] 24.4 Review error messages for information leakage
- [ ] 24.5 Test TLS certificate validation
- [ ] 24.6 Document security best practices
- [ ] 24.7 Set up security policy (SECURITY.md)
- [ ] 24.8 Configure dependabot security alerts

## 25. Community & Maintenance

- [ ] 25.1 Create CONTRIBUTING.md
- [ ] 25.2 Create issue templates
- [ ] 25.3 Create pull request template
- [ ] 25.4 Set up code owners (CODEOWNERS)
- [ ] 25.5 Create support/discussion channels
- [ ] 25.6 Document maintenance procedures
- [ ] 25.7 Plan for community engagement
