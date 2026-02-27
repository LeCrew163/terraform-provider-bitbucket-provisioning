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
