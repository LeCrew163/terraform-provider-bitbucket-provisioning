terraform {
  required_providers {
    bitbucketdc = {
      source  = "colab.internal.sldo.cloud/alpina/bitbucket-dc"
      version = "~> 0.1"
    }
  }
}

# Provider reads credentials from environment variables when attributes are omitted:
#   BITBUCKET_BASE_URL, BITBUCKET_USERNAME, BITBUCKET_PASSWORD (or BITBUCKET_TOKEN)
provider "bitbucketdc" {
  base_url             = var.bitbucket_url
  username             = var.bitbucket_username
  password             = var.bitbucket_password
  insecure_skip_verify = var.insecure_skip_verify
}

# ── Test: create a private project ─────────────────────────────────────────
resource "bitbucketdc_project" "test_private" {
  key         = "TFLOCAL"
  name        = "Terraform Local Test"
  description = "Private project created by the Terraform provider local test"
  public      = false
}

# ── Test: create a public project ──────────────────────────────────────────
resource "bitbucketdc_project" "test_public" {
  key    = "TFLPUB"
  name   = "Terraform Local Test Public"
  public = true
}

# ── Test: project with no optional fields ──────────────────────────────────
resource "bitbucketdc_project" "test_minimal" {
  key  = "TFLMIN"
  name = "Terraform Local Test Minimal"
}

# ── Test: repository within the private project ────────────────────────────
resource "bitbucketdc_repository" "api" {
  project_key = bitbucketdc_project.test_private.key
  name        = "API Service"
  description = "Main API service repository"
  forkable    = true
  public      = false
}

resource "bitbucketdc_repository" "frontend" {
  project_key = bitbucketdc_project.test_private.key
  name        = "Frontend App"
}

# ── Test: project-level permissions ───────────────────────────────────────
resource "bitbucketdc_project_permissions" "test_private" {
  project_key = bitbucketdc_project.test_private.key

  user {
    name       = "admin"
    permission = "PROJECT_ADMIN"
  }
}

# ── Test: branch restrictions on the private project ──────────────────────
resource "bitbucketdc_branch_permissions" "test_private" {
  project_key = bitbucketdc_project.test_private.key

  restriction {
    type         = "no-deletes"
    matcher_type = "BRANCH"
    matcher_id   = "main"
  }

  restriction {
    type         = "fast-forward-only"
    matcher_type = "PATTERN"
    matcher_id   = "release/*"
  }
}

# ── Test: project SSH access key ──────────────────────────────────────────
resource "bitbucketdc_project_access_key" "ci_pipeline" {
  project_key = bitbucketdc_project.test_private.key
  public_key  = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHj27twPopsFRMno1cyptyxbcKG3zSlFq7oCWPzWzlyP test-ci-pipeline"
  label       = "CI Pipeline Read Key"
  permission  = "PROJECT_READ"
}

# ── Test: repository-level permissions ────────────────────────────────────
resource "bitbucketdc_repository_permissions" "api" {
  project_key     = bitbucketdc_project.test_private.key
  repository_slug = bitbucketdc_repository.api.slug

  user {
    name       = "admin"
    permission = "REPO_ADMIN"
  }
}

# ── Test: repository SSH access key ───────────────────────────────────────
resource "bitbucketdc_repository_access_key" "frontend_deploy" {
  project_key     = bitbucketdc_project.test_private.key
  repository_slug = bitbucketdc_repository.frontend.slug
  public_key      = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOg1UJ5c9DFlej9e6jLiv3X8YL4e0eCH75dwG+VrtP8q test-deploy-key"
  label           = "Frontend Deploy Key"
  permission      = "REPO_READ"
}

# ── Repository hook (generic plugin hook management) ──────────────────────
# Example: enable the built-in "Reject Force Push" hook.
# Replace hook_key and settings_json for any other plugin hook
# (e.g. Webhook to Jenkins for Bitbucket Server).
resource "bitbucketdc_repository_hook" "force_push_protection" {
  project_key     = bitbucketdc_project.test_private.key
  repository_slug = bitbucketdc_repository.api.slug
  hook_key        = "com.atlassian.bitbucket.server.bitbucket-bundled-hooks:force-push-hook"
  enabled         = true
  settings_json   = jsonencode({})
}

# ── Webhook resources ─────────────────────────────────────────────────────
resource "bitbucketdc_webhook" "repo_hook" {
  project_key              = bitbucketdc_project.test_private.key
  repository_slug          = bitbucketdc_repository.api.slug
  name                     = "CI Notification Hook"
  url                      = "http://example.com/ci-hook"
  events                   = ["repo:refs_changed", "pr:opened"]
  active                   = true
  ssl_verification_required = false
}

resource "bitbucketdc_webhook" "project_hook" {
  project_key = bitbucketdc_project.test_private.key
  name        = "Project Audit Hook"
  url         = "http://example.com/audit-hook"
  events      = ["repo:refs_changed"]
}

# ── Default reviewer conditions ───────────────────────────────────────────
resource "bitbucketdc_default_reviewers" "test_private" {
  project_key = bitbucketdc_project.test_private.key

  condition {
    source_matcher_type = "ANY_REF"
    source_matcher_id   = "ANY_REF_MATCHER_ID"
    target_matcher_type = "ANY_REF"
    target_matcher_id   = "ANY_REF_MATCHER_ID"
    users               = ["admin"]
    required_approvals  = 1
  }
}

# ── Data sources ──────────────────────────────────────────────────────────
data "bitbucketdc_project" "test_private" {
  key        = bitbucketdc_project.test_private.key
  depends_on = [bitbucketdc_project.test_private]
}

data "bitbucketdc_repository" "api" {
  project_key = bitbucketdc_project.test_private.key
  slug        = bitbucketdc_repository.api.slug
  depends_on  = [bitbucketdc_repository.api]
}

data "bitbucketdc_user" "admin" {
  slug = "admin"
}

data "bitbucketdc_group" "stash_users" {
  name = "stash-users"
}

data "bitbucketdc_projects" "all" {
  depends_on = [bitbucketdc_project.test_private]
}

data "bitbucketdc_repositories" "all_in_private" {
  project_key = bitbucketdc_project.test_private.key
  depends_on  = [bitbucketdc_repository.api]
}

# ── Outputs ────────────────────────────────────────────────────────────────
output "private_project_id" {
  description = "Numeric ID of the private test project"
  value       = bitbucketdc_project.test_private.id
}

output "private_project_key" {
  description = "Key of the private test project"
  value       = bitbucketdc_project.test_private.key
}

output "public_project_id" {
  description = "Numeric ID of the public test project"
  value       = bitbucketdc_project.test_public.id
}

output "minimal_project_id" {
  description = "Numeric ID of the minimal test project"
  value       = bitbucketdc_project.test_minimal.id
}

output "api_repo_slug" {
  description = "Slug of the API service repository (derived from name by Bitbucket)"
  value       = bitbucketdc_repository.api.slug
}

output "api_repo_clone_http" {
  description = "HTTP clone URL for the API service repository"
  value       = bitbucketdc_repository.api.clone_url_http
}

output "frontend_repo_slug" {
  description = "Slug of the frontend repository"
  value       = bitbucketdc_repository.frontend.slug
}

output "ci_pipeline_key_id" {
  description = "Numeric ID of the CI Pipeline project access key"
  value       = bitbucketdc_project_access_key.ci_pipeline.key_id
}

output "ci_pipeline_key_fingerprint" {
  description = "Fingerprint of the CI Pipeline project access key"
  value       = bitbucketdc_project_access_key.ci_pipeline.fingerprint
}

output "ds_project_name" {
  description = "Project name read back via data source"
  value       = data.bitbucketdc_project.test_private.name
}

output "ds_repo_state" {
  description = "Repository state read back via data source"
  value       = data.bitbucketdc_repository.api.state
}

output "ds_admin_display_name" {
  description = "Display name of the admin user"
  value       = data.bitbucketdc_user.admin.display_name
}

output "ds_group_name" {
  description = "Group name read back via data source"
  value       = data.bitbucketdc_group.stash_users.name
}

output "repo_webhook_id" {
  description = "Numeric ID of the repository CI webhook"
  value       = bitbucketdc_webhook.repo_hook.webhook_id
}

output "project_webhook_id" {
  description = "Numeric ID of the project audit webhook"
  value       = bitbucketdc_webhook.project_hook.webhook_id
}

output "default_reviewers_id" {
  description = "ID of the default reviewer conditions resource"
  value       = bitbucketdc_default_reviewers.test_private.id
}

output "force_push_hook_enabled" {
  description = "Whether the force-push protection hook is enabled"
  value       = bitbucketdc_repository_hook.force_push_protection.enabled
}

output "ds_projects_count" {
  description = "Number of projects returned by the projects data source"
  value       = length(data.bitbucketdc_projects.all.projects)
}

output "ds_repos_count" {
  description = "Number of repositories in the private project"
  value       = length(data.bitbucketdc_repositories.all_in_private.repositories)
}
