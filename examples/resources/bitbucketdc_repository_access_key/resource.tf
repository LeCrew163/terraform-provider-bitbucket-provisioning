resource "bitbucketdc_repository_access_key" "ci" {
  project_key     = "PLAT"
  repository_slug = "platform-api"
  public_key      = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... ci-deploy-key"
  label           = "CI Deploy Key"
  permission      = "REPO_READ"
}
