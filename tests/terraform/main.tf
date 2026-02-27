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
  name   = "Terraform Local Test (Public)"
  public = true
}

# ── Test: project with no optional fields ──────────────────────────────────
resource "bitbucketdc_project" "test_minimal" {
  key  = "TFLMIN"
  name = "Terraform Local Test (Minimal)"
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
