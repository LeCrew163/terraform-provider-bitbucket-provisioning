# Project-scoped default reviewers
resource "bitbucketdc_default_reviewers" "platform" {
  project_key = "PLAT"

  condition {
    source_matcher_type = "ANY_REF"
    source_matcher_id   = "any"
    target_matcher_type = "BRANCH"
    target_matcher_id   = "main"
    users               = ["alice", "bob"]
    required_approvals  = 1
  }
}

# Repository-scoped default reviewers
resource "bitbucketdc_default_reviewers" "api" {
  project_key     = "PLAT"
  repository_slug = "platform-api"

  condition {
    source_matcher_type = "PATTERN"
    source_matcher_id   = "feature/*"
    target_matcher_type = "BRANCH"
    target_matcher_id   = "main"
    users               = ["alice"]
    required_approvals  = 1
  }
}
