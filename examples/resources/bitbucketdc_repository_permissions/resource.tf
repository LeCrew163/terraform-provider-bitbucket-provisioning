resource "bitbucketdc_repository_permissions" "api" {
  project_key     = "PLAT"
  repository_slug = "platform-api"

  group {
    name       = "platform-developers"
    permission = "REPO_WRITE"
  }

  group {
    name       = "platform-leads"
    permission = "REPO_ADMIN"
  }

  user {
    name       = "ci-bot"
    permission = "REPO_READ"
  }
}
