data "bitbucketdc_repository" "api" {
  project_key = "PLAT"
  slug        = "platform-api"
}

output "clone_url_ssh" {
  value = data.bitbucketdc_repository.api.clone_url_ssh
}
