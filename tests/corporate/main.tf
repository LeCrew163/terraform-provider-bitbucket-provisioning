terraform {
  required_version = "~> 1.5"
  required_providers {
    bitbucketdc = {
      source  = "LeCrew163/bitbucket-provisioning"
      version = "= 0.10.4"
    }
  }
}

# Credentials come entirely from environment variables:
#   BITBUCKET_BASE_URL=https://bitbucket.example.com
#   BITBUCKET_TOKEN=<token>
provider "bitbucketdc" {}

# ── Import an existing project ────────────────────────────────────────────
import {
  to = bitbucketdc_project.sandbox
  id = "MY-PROJECT"
}

resource "bitbucketdc_project" "sandbox" {
  key         = "MY-PROJECT"
  name        = "my-project"
  description = "Sandbox"
}

# ── Test repo ──────────────────────────────────────────────────────────────
resource "bitbucketdc_repository" "test" {
  project_key     = bitbucketdc_project.sandbox.key
  name            = "tf-provider-test"
  description     = "Temporary repo for Terraform provider prevent_destroy test"
  prevent_destroy = false
}
