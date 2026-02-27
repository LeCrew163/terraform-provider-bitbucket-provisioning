# Changelog

All notable changes to the Bitbucket Data Center Terraform Provider are documented here.

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
Versioning follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

## [0.1.0] — 2026-02-27

### Added

#### `bitbucketdc_project` resource
- Full CRUD lifecycle management for Bitbucket DC projects
- Attributes: `key` (required, immutable), `name`, `description` (optional), `public` (optional, computed)
- Plan-time key validation: uppercase letters, digits and underscores only, must start with a letter, 2–128 chars
- Import by project key: `terraform import bitbucketdc_project.name KEY`
- Acceptance tests: basic lifecycle, minimal config, key replace, disappears/drift, duplicate key, invalid key plans

#### `bitbucketdc_repository` resource
- Full CRUD lifecycle management for Bitbucket DC repositories within a project
- Attributes: `project_key` (required, immutable), `name` (required), `description` (optional), `forkable` (optional), `public` (optional)
- Computed attributes: `slug` (Bitbucket derives from name), `id`, `clone_url_http`, `clone_url_ssh`, `default_branch`, `state`
- Slug is server-derived from the repository name (Bitbucket DC ignores explicit slug on create)
- When name changes on update, Bitbucket also renames the slug — provider reads back the new slug
- Import by `PROJECT_KEY/slug`: `terraform import bitbucketdc_repository.name PROJECT_KEY/my-repo`
- Acceptance tests: basic lifecycle with update, minimal config, import

#### `bitbucketdc_project_permissions` resource
- Manages user and group permissions at the project level via reconciliation
- Nested blocks: `user { name, permission }` and `group { name, permission }`
- Permission levels: `PROJECT_READ`, `PROJECT_WRITE`, `PROJECT_ADMIN`
- Plan-time permission level validation
- Reconciliation: grants missing permissions, updates changed ones, revokes removed ones
- Import by project key: `terraform import bitbucketdc_project_permissions.name KEY`
- Acceptance tests: basic lifecycle (create, import, change level, empty config), invalid permission plan

#### `bitbucketdc_branch_permissions` resource
- Manages project-level branch restriction rules via reconciliation
- Nested blocks: `restriction { type, matcher_type, matcher_id, users, groups }`
- Restriction types: `read-only`, `no-deletes`, `fast-forward-only`, `pull-request-only`
- Matcher types: `BRANCH`, `PATTERN`, `MODEL_CATEGORY`, `MODEL_BRANCH`, `ANY_REF`
- Plan-time validation for both type and matcher_type values
- Reconciliation by semantic key `(type, matcher_type, matcher_id)`: deletes removed rules, creates new ones, replaces changed ones
- Handles Bitbucket API quirk: POST returns JSON array but generated client expects single object — treats 2xx as success
- Normalises empty user/group sets to null in state to prevent perpetual plan diffs inside SetNestedBlock
- Import by project key: `terraform import bitbucketdc_branch_permissions.name KEY`
- Acceptance tests: basic lifecycle (create, import, add rule, remove rule, empty), invalid type/matcher plan-only tests

#### Provider
- Authentication: Personal Access Token (`BITBUCKET_TOKEN`) or HTTP Basic Auth (`BITBUCKET_USERNAME` / `BITBUCKET_PASSWORD`)
- Configuration: `base_url`, `token`, `username`, `password`, `insecure_skip_verify`, `timeout`
- All provider attributes can be set via environment variables
- Configured via Terraform Plugin Framework (Protocol v6)

#### Infrastructure
- OpenAPI-generated Go client from Bitbucket DC 10.0 spec
- `Makefile` targets: `build`, `install`, `test`, `testacc`, `test-local`, `docker-*`
- Docker Compose setup for local Bitbucket DC instance
- End-to-end test script (`scripts/test-local.sh`) with plan/apply/drift-check/destroy cycle
- Terraform test configuration (`tests/terraform/`) covering all four resource types

[0.1.0]: https://bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/commits/tag/v0.1.0
