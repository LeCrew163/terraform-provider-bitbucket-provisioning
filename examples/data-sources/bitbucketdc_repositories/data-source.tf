# List all repositories in a project
data "bitbucketdc_repositories" "all" {
  project_key = "PLAT"
}

# List only repositories whose name contains "api"
data "bitbucketdc_repositories" "apis" {
  project_key = "PLAT"
  filter      = "api"
}

output "clone_urls" {
  value = [for r in data.bitbucketdc_repositories.all.repositories : r.clone_url_ssh]
}
