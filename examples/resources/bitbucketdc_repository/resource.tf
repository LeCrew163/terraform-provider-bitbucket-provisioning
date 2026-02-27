resource "bitbucketdc_repository" "api" {
  project_key = "PLAT"
  name        = "platform-api"
  description = "REST API service"
  forkable    = true
  public      = false
}
