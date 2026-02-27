resource "bitbucketdc_branch_permissions" "platform" {
  project_key = "PLAT"

  # Protect main — only allow merges via pull request
  restriction {
    type         = "pull-request-only"
    matcher_type = "BRANCH"
    matcher_id   = "main"
    groups       = ["platform-leads"]
  }

  # Prevent force-pushes on release branches
  restriction {
    type         = "fast-forward-only"
    matcher_type = "PATTERN"
    matcher_id   = "release/*"
  }

  # No one may delete tags
  restriction {
    type         = "no-deletes"
    matcher_type = "ANY_REF"
    matcher_id   = "any"
  }
}
