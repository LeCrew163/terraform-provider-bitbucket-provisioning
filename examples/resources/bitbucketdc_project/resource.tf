resource "bitbucketdc_project" "platform" {
  key         = "PLAT"
  name        = "Platform Team"
  description = "Shared platform infrastructure repositories"
  public      = false
}
