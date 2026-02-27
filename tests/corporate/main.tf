terraform {
  required_providers {
    bitbucketdc = {
      source  = "art01.sldnet.de:8081/swisslife/bitbucket-provisioning"
      version = "= 0.10.0"
    }
  }
}

# Credentials come entirely from environment variables:
#   BITBUCKET_BASE_URL=https://bitbucket.colab.internal.sldo.cloud
#   BITBUCKET_TOKEN=<token>
provider "bitbucketdc" {}

# ── Import existing ALPINA-SANDBOX project ────────────────────────────────
import {
  to = bitbucketdc_project.sandbox
  id = "ALPINA-SANDBOX"
}

resource "bitbucketdc_project" "sandbox" {
  key         = "ALPINA-SANDBOX"
  name        = "alpina-sandbox"
  description = "Sandbox"
}

# ── Test repo ──────────────────────────────────────────────────────────────
resource "bitbucketdc_repository" "test" {
  project_key     = bitbucketdc_project.sandbox.key
  name            = "tf-provider-test"
  description     = "Temporary repo for Terraform provider prevent_destroy test"
  prevent_destroy = false
}
