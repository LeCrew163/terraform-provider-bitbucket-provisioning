# Changelog

All notable changes to the Bitbucket Data Center Terraform Provider are documented here.

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
Versioning follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

## [0.10.0] — 2026-02-27

### Added

#### Provider Documentation — usage examples
- Added `## Example Usage` sections to all resource and data source docs
- `examples/resources/bitbucketdc_*/resource.tf` — realistic usage examples for all 11 resources
- `examples/data-sources/bitbucketdc_*/data-source.tf` — usage examples for all 6 data sources
- `docs/` directory is now tracked in version control (removed from `.gitignore`)
- `make docs` target updated to use `go run` so no local `tfplugindocs` install is required

## [0.9.0] — 2026-02-27

### Added

#### CI/CD — `jenkins/` pipeline folder
- `jenkins/ci.groovy` — Build, Lint, and Unit Tests; runs on every commit
- `jenkins/cd.groovy` — Build and Acceptance Tests against a live Bitbucket instance (parameterised via `RUN_ACC_TESTS`); requires Jenkins credentials `bitbucket-base-url`, `bitbucket-username`, `bitbucket-password`
- `jenkins/release.groovy` — Build and GoReleaser publish; requires `github-token` credential

#### Provider Documentation
- Generated `docs/` directory using `tfplugindocs` — includes `index.md`, resource docs, and data source docs
- `make docs` target in Makefile uses `tfplugindocs generate --provider-name bitbucketdc`
- `templates/index.md.tmpl` provides the provider overview page

## [0.8.0] — 2026-02-27

### Added

#### `data.bitbucketdc_projects` data source
- Lists all Bitbucket Data Center projects accessible to the authenticated user
- Optional `filter` attribute for case-insensitive name substring matching
- Full pagination support — fetches all pages automatically
- Exposes per-project attributes: `id`, `key`, `name`, `description`, `public`

#### `data.bitbucketdc_repositories` data source
- Lists all repositories within a given Bitbucket Data Center project
- Required `project_key` attribute; optional `filter` for case-insensitive name substring matching
- Full pagination support — fetches all pages automatically
- Returns 404 error if the specified project does not exist
- Exposes per-repository attributes: `id`, `slug`, `name`, `description`, `public`, `state`, `forkable`, `default_branch`, `clone_url_http`, `clone_url_ssh`

## [0.7.0] — 2026-02-27

### Added

#### `bitbucketdc_project_hook` resource
- Generic resource for managing any Bitbucket Data Center **project-level** plugin hook (WORKFLOW → Hooks on the project)
- Project hooks apply to all repositories within the project — the primary use-case for Webhook to Jenkins
- Same interface as `bitbucketdc_repository_hook`: `project_key`, `hook_key`, `enabled`, `settings_json`
- Import by `project_key/hook_key`

---

## [0.6.0] — 2026-02-27

### Added

#### `bitbucketdc_repository_hook` resource
- Generic resource for managing any Bitbucket Data Center repository plugin hook (WORKFLOW → Hooks)
- Works with any plugin that uses the Bitbucket hook framework, including **Webhook to Jenkins for Bitbucket Server** (`com.nerdwin15.stash-stash-webhook-jenkins:jenkinsPostReceiveHook`) and all built-in hooks
- Attributes: `project_key`, `repository_slug`, `hook_key` (all `RequiresReplace`), `enabled` (default `true`), `settings_json`
- `settings_json` accepts arbitrary JSON via `jsonencode()` — the exact fields are plugin-specific (e.g. `jenkinsBase`, `cloneType` for the Jenkins hook)
- Settings are stored and compared as normalised (compact) JSON to avoid whitespace drift
- Import by `project_key/repository_slug/hook_key`
- Distinct from `bitbucketdc_webhook` which manages native Bitbucket webhooks (WORKFLOW → Webhooks)

---

## [0.5.0] — 2026-02-27

### Added

#### `bitbucketdc_webhook` resource
- Manages Bitbucket DC native webhooks (WORKFLOW → Webhooks) at project or repository scope
- Attributes: `project_key`, `repository_slug` (optional — omit for project scope), `name`, `url`, `events` (set), `active` (default `true`), `ssl_verification_required` (default `true`), `webhook_id` (computed)
- Events stored as a set (unordered) to avoid ordering drift when Bitbucket returns events in a different order
- Import by `project_key/webhook_id` (project scope) or `project_key/repository_slug/webhook_id` (repo scope)

#### `bitbucketdc_default_reviewers` resource
- Reconciliation resource for Bitbucket DC default reviewer conditions at project or repository scope
- Each `condition` block defines a rule that automatically adds reviewers to matching pull requests
- Condition attributes: `source_matcher_type`, `source_matcher_id`, `target_matcher_type`, `target_matcher_id`, `users` (list of usernames), `required_approvals`
- Conditions identified by semantic key `(source_type|source_id|target_type|target_id)` — no stored API IDs
- User slugs are resolved to full user objects (including numeric ID) via the SystemMaintenance API before being sent to Bitbucket
- Import by `project_key` or `project_key/repository_slug`

---

## [0.4.0] — 2026-02-27

### Added

#### `bitbucketdc_project` data source
- Reads an existing Bitbucket DC project by key
- Attributes: `key` (required), `id` (computed, equals key), `name`, `description`, `public`
- Returns a clear error when the project key is not found

#### `bitbucketdc_repository` data source
- Reads an existing Bitbucket DC repository by project key and slug
- Attributes: `project_key`, `slug` (both required), `id` (`{project_key}/{slug}`), `name`, `description`, `public`, `forkable`, `default_branch`, `clone_url_http`, `clone_url_ssh`, `state`
- Reuses the same clone URL extraction helper as the resource

#### `bitbucketdc_user` data source
- Reads an existing Bitbucket DC user by slug (username)
- Attributes: `slug` (required), `id` (computed, equals slug), `name`, `display_name`, `email_address`, `active`
- Returns a clear 404 error when the slug is not found

#### `bitbucketdc_group` data source
- Reads an existing Bitbucket DC group by exact name
- Attributes: `name` (required), `id` (computed, equals name)
- Uses the `GetGroups` list-with-filter endpoint and performs an exact-match check (Bitbucket DC has no get-by-name API)
- Returns a clear error when no exact match is found

---

## [0.3.0] — 2026-02-27

### Added

#### `bitbucketdc_repository_permissions` resource
- Full reconciliation lifecycle for user and group permissions at the repository level
- Attributes: `project_key` (required, immutable), `repository_slug` (required, immutable)
- Nested blocks: `user { name, permission }` and `group { name, permission }`
- Permission levels: `REPO_READ`, `REPO_WRITE`, `REPO_ADMIN`
- Reconciliation: grants missing permissions, updates changed ones, revokes removed ones
- Import by `project_key/repository_slug`: `terraform import bitbucketdc_repository_permissions.name PROJ/my-repo`
- Acceptance tests: basic lifecycle (create → import → update level → empty), invalid permission plan

#### `bitbucketdc_repository_access_key` resource
- Full CRUD lifecycle management for SSH access keys at the repository level
- Attributes: `project_key` (required, immutable), `repository_slug` (required, immutable), `public_key` (required, immutable), `label` (optional+computed, immutable), `permission` (required, updatable)
- Computed attributes: `key_id`, `fingerprint`, `id` (`{project_key}/{repository_slug}/{key_id}`)
- Permission levels: `REPO_READ`, `REPO_WRITE`
- Same 2xx+unmarshal recovery pattern as `bitbucketdc_project_access_key`
- Import by `project_key/repository_slug/key_id`: `terraform import bitbucketdc_repository_access_key.name PROJ/my-repo/42`
- Acceptance tests: basic lifecycle (create with label → import → update permission), no-label config, invalid permission plan

### Planned

#### `prevent_destroy` guard for destructive resources

`bitbucketdc_project` and `bitbucketdc_repository` resources should default to
safe-delete behaviour to prevent accidental data loss.

**Proposed design:**

- Add an optional `prevent_destroy` boolean attribute to `bitbucketdc_project`
  and `bitbucketdc_repository` (default `true`).
- When `prevent_destroy = true` the provider's `Delete` function returns an
  error immediately, blocking `terraform destroy` / resource replacement before
  it reaches the API.
- To destroy a resource the operator must:
  1. Set `prevent_destroy = false` in the configuration.
  2. Run `terraform apply` (updates state — no API call, attribute is
     provider-side only).
  3. Run `terraform destroy` (or trigger replacement).
- This is a **provider-level guard**, distinct from Terraform's built-in
  `lifecycle { prevent_destroy = true }` meta-argument (which lives only in
  config and can be silently bypassed by removing the block before running
  destroy).
- The attribute should be `Optional`, `Default: true`, and should NOT trigger
  resource replacement when toggled (use `planmodifier` that suppresses the
  diff without replace).
- On import the attribute should default to `true` so imported resources are
  protected immediately.
- Consider extending the same guard to `bitbucketdc_branch_permissions` and
  `bitbucketdc_project_permissions` if destruction is also considered risky
  there.

---

## [0.2.0] — 2026-02-27

### Added

#### `bitbucketdc_project_access_key` resource
- Full CRUD lifecycle management for SSH access keys at the Bitbucket DC project level
- Attributes: `project_key` (required, immutable), `public_key` (required, immutable), `label` (optional+computed, immutable), `permission` (required, updatable)
- Computed attributes: `key_id` (server-assigned numeric id), `fingerprint`, `id` (`{project_key}/{key_id}`)
- `label` is Optional+Computed: Bitbucket auto-derives a label from the SSH key comment when none is supplied
- Permission is updatable in-place via `UpdatePermission` API; all other changes force resource replacement
- Plan-time permission validation: only `PROJECT_READ` and `PROJECT_WRITE` are accepted
- Import by `project_key/key_id`: `terraform import bitbucketdc_project_access_key.name PROJ/42`
- Handles Bitbucket API quirk: generated client fails to unmarshal `AddForProject` response due to `DisallowUnknownFields` on nested type — recovers by listing keys and finding by public key text
- Root-cause fix: `AddSshKeyRequest.CreatedDate` changed from `*time.Time` to `*int64` — the API returns a Unix millisecond timestamp (integer), which Go's `time.Time` JSON unmarshaler rejects
- Acceptance tests: basic lifecycle (create with label → import → update permission), no-label config, invalid permission plan-only validation

### Fixed

- `internal/client/generated/model_add_ssh_key_request.go`: Changed `CreatedDate` field type from `*time.Time` to `*int64` to match the actual Bitbucket API response (Unix milliseconds), preventing JSON unmarshal failures across all access key read/list operations

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

[0.10.0]: https://bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/compare/v0.9.0..v0.10.0
[0.9.0]: https://bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/compare/v0.8.0..v0.9.0
[0.8.0]: https://bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/compare/v0.7.0..v0.8.0
[0.7.0]: https://bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/compare/v0.6.0..v0.7.0
[0.6.0]: https://bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/compare/v0.5.0..v0.6.0
[0.5.0]: https://bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/compare/v0.4.0..v0.5.0
[0.4.0]: https://bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/compare/v0.3.0..v0.4.0
[0.3.0]: https://bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/compare/v0.2.0..v0.3.0
[0.2.0]: https://bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/compare/v0.1.0..v0.2.0
[0.1.0]: https://bitbucket.colab.internal.sldo.cloud/alpina-operation/bitbucket-provisioning/commits/tag/v0.1.0
