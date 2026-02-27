# Repository-scoped webhook
resource "bitbucketdc_webhook" "ci_trigger" {
  project_key     = "PLAT"
  repository_slug = "platform-api"
  name            = "Jenkins CI Trigger"
  url             = "https://jenkins.example.com/bitbucket-hook/"
  active          = true
  events = [
    "repo:refs_changed",
    "pr:opened",
    "pr:merged",
  ]
  ssl_verification_required = true
}

# Project-scoped webhook (fires for all repos in the project)
resource "bitbucketdc_webhook" "audit" {
  project_key = "PLAT"
  name        = "Audit Webhook"
  url         = "https://audit.example.com/events"
  active      = true
  events      = ["repo:refs_changed", "pr:merged"]
}
