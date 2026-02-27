data "bitbucketdc_project" "platform" {
  key = "PLAT"
}

output "project_name" {
  value = data.bitbucketdc_project.platform.name
}
