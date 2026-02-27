# List all projects
data "bitbucketdc_projects" "all" {}

# List only projects whose name contains "platform"
data "bitbucketdc_projects" "platform" {
  filter = "platform"
}

output "project_keys" {
  value = [for p in data.bitbucketdc_projects.all.projects : p.key]
}
