terraform {
  required_version = "~> 1.5"
  required_providers {
    bitbucketdc = {
      source  = "art01.sldnet.de:8081/swisslife/bitbucket-provisioning"
      version = "~> 0.10"
    }
  }
}

# Authenticate using a Personal Access Token
provider "bitbucketdc" {
  base_url = "https://bitbucket.example.com"
  token    = var.bitbucket_token
}

# Or authenticate using username and password
provider "bitbucketdc" {
  base_url = "https://bitbucket.example.com"
  username = var.bitbucket_username
  password = var.bitbucket_password
}
