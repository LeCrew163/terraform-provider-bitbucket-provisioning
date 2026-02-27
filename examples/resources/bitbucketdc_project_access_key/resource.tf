resource "bitbucketdc_project_access_key" "ci" {
  project_key = "PLAT"
  public_key  = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... ci-deploy-key"
  label       = "CI Deploy Key"
  permission  = "PROJECT_READ"
}
